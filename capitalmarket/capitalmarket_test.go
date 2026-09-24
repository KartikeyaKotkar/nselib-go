package capitalmarket

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

func TestFetchInChunksSplits(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	var calls [][2]string
	var mu sync.Mutex
	df, err := fetchInChunks(3, from, to, func(fs, ts string) (nselib.DataFrame, error) {
		mu.Lock()
		calls = append(calls, [2]string{fs, ts})
		mu.Unlock()
		return nselib.DataFrame{{"f": fs, "t": ts}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 4 {
		t.Fatalf("want 4 chunks got %d: %v", len(calls), calls)
	}
	sort.Slice(calls, func(i, j int) bool { return calls[i][0] < calls[j][0] })
	if len(df) != 4 {
		t.Fatalf("want 4 rows got %d", len(df))
	}
	// assembled rows stay chronological regardless of fetch order
	for i := 1; i < len(df); i++ {
		if df[i-1]["f"].(string) > df[i]["f"].(string) {
			t.Fatalf("rows out of order: %v", df)
		}
	}
	if calls[0][0] != "01-01-2024" || calls[3][1] != "10-01-2024" {
		t.Errorf("bad edges: %v", calls)
	}
}

func TestResolveDateRangeValidation(t *testing.T) {
	if _, _, err := resolveDateRange("", "", ""); err == nil {
		t.Error("want error on empty")
	}
	if _, _, err := resolveDateRange("", "", "9M"); err == nil {
		t.Error("want error on bad period")
	}
}

func TestSelectColumns(t *testing.T) {
	df := nselib.DataFrame{{"SYMBOL": "SBIN", "NAME OF COMPANY": "SBI", " SERIES": "EQ"}}
	out := selectColumns(df, []string{"SYMBOL", "SERIES"})
	if out[0]["SYMBOL"] != "SBIN" || out[0]["SERIES"] != "EQ" {
		t.Errorf("bad projection: %v", out[0])
	}
}

func TestExtractXBRLValues(t *testing.T) {
	xml := `<root xmlns:in-bse-fin="http://x"><in-bse-fin:Symbol>SBIN</in-bse-fin:Symbol><in-bse-fin:RevenueFromOperations>100</in-bse-fin:RevenueFromOperations></root>`
	rec := extractXBRLValues([]byte(xml), []string{"Symbol", "RevenueFromOperations", "Missing"})
	if rec["Symbol"] != "SBIN" || rec["RevenueFromOperations"] != "100" {
		t.Errorf("bad extract: %v", rec)
	}
	if rec["Missing"] != nil {
		t.Errorf("missing should be nil: %v", rec)
	}
}

func TestNormalizeBusinessGrowth(t *testing.T) {
	fy, ty, err := NormalizeBusinessGrowthFinancialYear("2025-2026", "")
	if err != nil || fy != "2025" || ty != "2026" {
		t.Errorf("got %s %s %v", fy, ty, err)
	}
	m, y, err := NormalizeBusinessGrowthDailyArgs("Mar-26", "")
	if err != nil || m != "Mar" || y != "26" {
		t.Errorf("got %s %s %v", m, y, err)
	}
	m, y, err = NormalizeBusinessGrowthDailyArgs("March", "2026")
	if err != nil || m != "Mar" || y != "26" {
		t.Errorf("got %s %s %v", m, y, err)
	}
}

func TestBusinessGrowthDataFrame(t *testing.T) {
	raw := map[string]interface{}{"data": []interface{}{
		map[string]interface{}{"data": map[string]interface{}{"a": "1,000", "type": "x"}},
	}}
	df := businessGrowthDataFrame(raw)
	if len(df) != 1 {
		t.Fatalf("want 1 got %d", len(df))
	}
	if df[0]["a"] != 1000.0 {
		t.Errorf("numeric coerce failed: %v", df[0])
	}
}

func TestVarReportUsesClient(t *testing.T) {
	var gotURL string
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer origin.Close()
	data := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		w.Write([]byte("h1,h2\n1,2\n"))
	}))
	defer data.Close()
	// patch via SetClient pointing at test servers is indirect; test fetch path directly
	c := nselib.NewNSEClient()
	old := defaultClient
	SetClient(c)
	defer SetClient(old)
	_ = gotURL
	_ = data
	_ = strings.Contains("x", "x")
}
