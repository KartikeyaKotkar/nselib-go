package nselib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// Browser-mimicking headers ported from nselib/libutil.py.
var (
	defaultOriginHeader = map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
	}
	dataFetchHeader = map[string]string{
		"referer":                "https://www.nseindia.com/",
		"Connection":             "keep-alive",
		"Cache-Control":          "max-age=0",
		"DNT":                    "1",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
		"Sec-Fetch-User":         "?1",
		"Accept":                 "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"Sec-Fetch-Site":         "none",
		"Sec-Fetch-Mode":         "navigate",
		"Accept-Language":        "en-US,en;q=0.9,hi;q=0.8",
	}
)

// DefaultOriginURL mirrors Python default origin for cookie priming.
const DefaultOriginURL = "https://www.nseindia.com"

// NSEClient manages HTTP sessions with cookie priming for NSE India APIs.
type NSEClient struct {
	client *http.Client
	logger *slog.Logger
}

// ClientOption configures NSEClient.
type ClientOption func(*NSEClient)

// WithHTTPClient overrides the internal http.Client.
func WithHTTPClient(c *http.Client) ClientOption {
	return func(n *NSEClient) { n.client = c }
}

// WithLogger overrides the logger.
func WithLogger(l *slog.Logger) ClientOption {
	return func(n *NSEClient) { n.logger = l }
}

// NewNSEClient creates a client with a cookie jar and standard timeouts.
func NewNSEClient(opts ...ClientOption) *NSEClient {
	jar, _ := cookiejar.New(nil)
	c := &NSEClient{
		client: &http.Client{Jar: jar, Timeout: 30 * time.Second},
		logger: slog.Default(),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func applyHeaders(req *http.Request, h map[string]string) {
	for k, v := range h {
		req.Header.Set(k, v)
	}
}

// primeCookies performs GET originURL to seed the cookie jar.
func (c *NSEClient) primeCookies(originURL string) error {
	req, err := http.NewRequest(http.MethodGet, originURL, nil)
	if err != nil {
		return err
	}
	applyHeaders(req, defaultOriginHeader)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	return nil
}

// Fetch primes cookies from originURL, then fetches target URL.
// On 403 it re-primes once and retries (cookie expiry recovery).
// Equivalent to Python's nse_urlfetch().
func (c *NSEClient) Fetch(url, originURL string) (*http.Response, error) {
	if originURL == "" {
		originURL = DefaultOriginURL
	}
	if err := c.primeCookies(originURL); err != nil {
		return nil, fmt.Errorf("prime cookies: %w", err)
	}
	doFetch := func() (*http.Response, error) {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		applyHeaders(req, dataFetchHeader)
		return c.client.Do(req)
	}
	resp, err := doFetch()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusForbidden {
		_ = resp.Body.Close()
		if err := c.primeCookies(originURL); err != nil {
			return nil, fmt.Errorf("re-prime cookies: %w", err)
		}
		resp, err = doFetch()
		if err != nil {
			return nil, err
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		return nil, NewAPIError(fmt.Sprintf("GET %s: status %d: %s", url, resp.StatusCode, string(body)))
	}
	return resp, nil
}

// FetchBytes fetches URL and returns raw body bytes.
func (c *NSEClient) FetchBytes(url, originURL string) ([]byte, error) {
	resp, err := c.Fetch(url, originURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// FetchJSON decodes JSON response into target.
func (c *NSEClient) FetchJSON(url, originURL string, target interface{}) error {
	body, err := c.FetchBytes(url, originURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, target); err != nil {
		return NewAPIError(fmt.Sprintf("decode JSON from %s: %v", url, err))
	}
	return nil
}

// FetchCSV fetches URL and parses CSV body into DataFrame.
func (c *NSEClient) FetchCSV(url, originURL string, opts ...CSVOption) (DataFrame, error) {
	body, err := c.FetchBytes(url, originURL)
	if err != nil {
		return nil, err
	}
	return ParseCSV(bytes.NewReader(body), opts...)
}
