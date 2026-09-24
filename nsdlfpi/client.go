package nsdlfpi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

// HTMLProvider supplies rendered page HTML (browser automation).
type HTMLProvider interface {
	LatestHTML() (string, error)
	ArchiveHTML(tradeDate time.Time) (string, error)
	Available() bool
}

// Client fetches NSDL FPI reports with pilot/production failover.
type Client struct {
	http     *http.Client
	provider HTMLProvider
	baseURL  string
	fields   map[string]string
}

// NewClient creates a client; provider may be nil (request fallback only).
func NewClient(provider HTMLProvider) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		http:     &http.Client{Jar: jar, Timeout: 60 * time.Second},
		provider: provider,
	}
}

func (c *Client) get(page string) (string, error) {
	resp, err := c.http.Get(c.baseURL + "/" + page)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

// resolveBaseURL tries pilot then production like Python's REQUEST_BASE_URLS.
func (c *Client) resolveBaseURL(page string) error {
	if c.baseURL != "" {
		return nil
	}
	var lastErr error
	for _, base := range []string{PilotBaseURL, ProductionBaseURL} {
		c.baseURL = base
		text, err := c.get(page)
		if err != nil {
			lastErr = err
			continue
		}
		if strings.Contains(text, "NSDL") && strings.Contains(text, "FPI") {
			if page == ArchivePage {
				c.fields = ExtractHiddenFields(text)
			}
			return nil
		}
	}
	c.baseURL = ""
	return fmt.Errorf("unable to reach NSDL FPI reports page: %v", lastErr)
}

func (c *Client) archiveFields(refresh bool) (map[string]string, error) {
	if c.fields != nil && !refresh {
		return c.fields, nil
	}
	if err := c.resolveBaseURL(ArchivePage); err != nil {
		return nil, err
	}
	text, err := c.get(ArchivePage)
	if err != nil {
		return nil, err
	}
	fields := ExtractHiddenFields(text)
	for _, v := range fields {
		if v == "" {
			return nil, fmt.Errorf("unable to prepare NSDL archive request")
		}
	}
	c.fields = fields
	return fields, nil
}

func (c *Client) latestHTML() (string, string, error) {
	var errs []string
	if c.provider != nil && c.provider.Available() {
		if html, err := c.provider.LatestHTML(); err == nil {
			return html, ProductionBaseURL, nil
		} else {
			errs = append(errs, "production_browser: "+err.Error())
		}
	}
	if err := c.resolveBaseURL(LatestPage); err != nil {
		errs = append(errs, "request_fallback: "+err.Error())
		return "", "", errors.New("unable to fetch NSDL latest page :: " + strings.Join(errs, " | "))
	}
	text, err := c.get(LatestPage)
	if err != nil {
		return "", "", fmt.Errorf("unable to fetch NSDL latest page :: request_fallback: %v", err)
	}
	return text, c.baseURL, nil
}

func (c *Client) archiveHTML(tradeDate time.Time) (string, string, error) {
	var errs []string
	if c.provider != nil && c.provider.Available() {
		if html, err := c.provider.ArchiveHTML(tradeDate); err == nil {
			return html, ProductionBaseURL, nil
		} else {
			errs = append(errs, "production_browser: "+err.Error())
		}
	}
	fields, err := c.archiveFields(false)
	if err != nil {
		errs = append(errs, "request_fallback: "+err.Error())
		return "", "", errors.New("unable to fetch NSDL archive page :: " + strings.Join(errs, " | "))
	}
	form := url.Values{
		"__EVENTTARGET":        {"btnSubmit1"},
		"__EVENTARGUMENT":      {""},
		"__VIEWSTATE":          {fields["__VIEWSTATE"]},
		"__VIEWSTATEGENERATOR": {fields["__VIEWSTATEGENERATOR"]},
		"__EVENTVALIDATION":    {fields["__EVENTVALIDATION"]},
		"txtDate":              {tradeDate.Format(ReportDateLayout)},
		"hdnDate":              {tradeDate.Format(ReportDateLayout)},
		"HdnValexceldata":      {""},
		"hdnFlag":              {""},
	}
	resp, err := c.http.PostForm(c.baseURL+"/"+ArchivePage, form)
	if err != nil {
		return "", "", fmt.Errorf("unable to fetch NSDL archive page :: request_fallback: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("unable to fetch NSDL archive page :: status %d", resp.StatusCode)
	}
	_, _ = c.archiveFields(true)
	return string(body), c.baseURL, nil
}

// LatestBundle fetches the latest report bundle.
func (c *Client) LatestBundle() (*ReportBundle, error) {
	html, base, err := c.latestHTML()
	if err != nil {
		return nil, err
	}
	source := "latest_request_fallback"
	if base == ProductionBaseURL && c.provider != nil && c.provider.Available() {
		source = "latest_production_browser"
	}
	b := ParseReportBundle(html, source, nil)
	b.AsOfDate = MaxBundleReportDate(b)
	if len(b.Investment) == 0 && len(b.Derivative) == 0 {
		return nil, nselib.NewDataNotFoundError("NSDL latest page did not contain investment or derivative data")
	}
	return b, nil
}

// ArchiveMonthBundle fetches the bundle for a month, walking back up to 10 days.
func (c *Client) ArchiveMonthBundle(tradeDate time.Time, maxLookbackDays int) (*ReportBundle, error) {
	monthKey := tradeDate.Format("2006-01")
	current := tradeDate
	var lastErr error
	for i := 0; i <= maxLookbackDays; i++ {
		if current.Format("2006-01") != monthKey {
			break
		}
		html, base, err := c.archiveHTML(current)
		if err != nil {
			lastErr = err
		} else if !strings.Contains(html, "No Data To Display") {
			source := "archive_request_fallback"
			if base == ProductionBaseURL && c.provider != nil && c.provider.Available() {
				source = "archive_production_browser"
			}
			asOf := current
			if b := ParseReportBundle(html, source, &asOf); len(b.Investment) > 0 || len(b.Derivative) > 0 {
				return b, nil
			}
		}
		current = current.AddDate(0, 0, -1)
	}
	if lastErr != nil {
		return nil, fmt.Errorf("NSDL archive data not available for month ending %s: %v",
			tradeDate.Format(ReportDateLayout), lastErr)
	}
	return nil, fmt.Errorf("NSDL archive data not available for month ending %s",
		tradeDate.Format(ReportDateLayout))
}

// defaultClient mirrors Python's per-call NSDLFPIClient with shared browser.
var defaultClient = NewClient(NewBrowser())

// FetchLatestBundle returns the latest NSDL FPI bundle.
func FetchLatestBundle() (*ReportBundle, error) { return defaultClient.LatestBundle() }

// FetchMonthBundle returns the bundle for the month containing tradeDate (dd-mm-YYYY).
func FetchMonthBundle(tradeDate string) (*ReportBundle, error) {
	t, err := CoerceTradeDate(tradeDate)
	if err != nil {
		return nil, err
	}
	return defaultClient.ArchiveMonthBundle(t, 10)
}

// FetchLatestInvestmentActivity returns latest investment rows.
func FetchLatestInvestmentActivity() (nselib.DataFrame, error) {
	b, err := FetchLatestBundle()
	if err != nil {
		return nil, err
	}
	return b.Investment, nil
}

// FetchLatestDerivativeActivity returns latest derivative rows.
func FetchLatestDerivativeActivity() (nselib.DataFrame, error) {
	b, err := FetchLatestBundle()
	if err != nil {
		return nil, err
	}
	return b.Derivative, nil
}

func filterByDate(df nselib.DataFrame, want string) nselib.DataFrame {
	var out nselib.DataFrame
	for _, r := range df {
		if s, _ := r["REPORT_DATE"].(string); s == want {
			out = append(out, r)
		}
	}
	if out == nil {
		out = nselib.DataFrame{}
	}
	return out
}

// FetchInvestmentActivity returns investment rows for an exact trade date.
func FetchInvestmentActivity(tradeDate string) (nselib.DataFrame, error) {
	t, err := CoerceTradeDate(tradeDate)
	if err != nil {
		return nil, err
	}
	b, err := defaultClient.ArchiveMonthBundle(t, 10)
	if err != nil {
		return nil, err
	}
	out := filterByDate(b.Investment, t.Format(ReportDateLayout))
	if len(out) == 0 {
		return nil, nselib.NewDataNotFoundError("NSDL investment activity not found for " + t.Format(ReportDateLayout))
	}
	return out, nil
}

// FetchDerivativeActivity returns derivative rows for an exact trade date.
func FetchDerivativeActivity(tradeDate string) (nselib.DataFrame, error) {
	t, err := CoerceTradeDate(tradeDate)
	if err != nil {
		return nil, err
	}
	b, err := defaultClient.ArchiveMonthBundle(t, 10)
	if err != nil {
		return nil, err
	}
	out := filterByDate(b.Derivative, t.Format(ReportDateLayout))
	if len(out) == 0 {
		return nil, nselib.NewDataNotFoundError("NSDL derivative activity not found for " + t.Format(ReportDateLayout))
	}
	return out, nil
}
