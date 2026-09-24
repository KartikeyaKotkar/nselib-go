package derivatives

import (
	"sync"
	"testing"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

func TestFetchInChunks90(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 4, 5, 0, 0, 0, 0, time.UTC) // 95 days -> 2 chunks of 90
	var calls [][2]string
	var mu sync.Mutex
	df, err := fetchInChunks(90, from, to, func(fs, ts string) (nselib.DataFrame, error) {
		mu.Lock()
		calls = append(calls, [2]string{fs, ts})
		mu.Unlock()
		return nselib.DataFrame{{"f": fs}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 {
		t.Fatalf("want 2 chunks got %v", calls)
	}
	if len(df) != 2 {
		t.Fatalf("want 2 rows got %d", len(df))
	}
}

func TestInstrumentValidation(t *testing.T) {
	if _, err := FuturePriceVolumeData("SBIN", "OPTIDX", "01-01-2024", "05-01-2024", ""); err == nil {
		t.Error("want error on wrong future instrument")
	}
	if _, err := OptionPriceVolumeData("NIFTY", "FUTSTK", "CE", "01-01-2024", "05-01-2024", ""); err == nil {
		t.Error("want error on wrong option instrument")
	}
	if _, err := OptionPriceVolumeData("NIFTY", "OPTIDX", "XX", "01-01-2024", "05-01-2024", ""); err == nil {
		t.Error("want error on bad option type")
	}
	if _, err := NSELiveOptionChain("NIFTY", "01-01-2024", "huge"); err == nil {
		t.Error("want error on bad oi mode")
	}
}

func TestParseOptionChain(t *testing.T) {
	raw := map[string]interface{}{
		"records": map[string]interface{}{
			"timestamp": "01-Jan-2024 15:30:00",
			"data": []interface{}{
				map[string]interface{}{
					"strikePrice": 22000.0, "expiryDate": "25-Jan-2024",
					"CE": map[string]interface{}{"openInterest": 100.0, "lastPrice": 50.0},
					"PE": map[string]interface{}{"openInterest": 200.0, "lastPrice": 60.0},
				},
				map[string]interface{}{
					"strikePrice": 22100.0, "expiryDate": "01-Feb-2024",
				},
			},
		},
	}
	full := parseOptionChain(raw, "NIFTY", "", "full")
	if len(full) != 2 {
		t.Fatalf("want 2 rows got %d", len(full))
	}
	if full[0]["CALLS_OI"] != 100.0 || full[0]["PUTS_OI"] != 200.0 {
		t.Errorf("bad values: %v", full[0])
	}
	if full[1]["CALLS_OI"] != 0 || full[1]["PUTS_OI"] != 0 {
		t.Errorf("missing side should default 0: %v", full[1])
	}
	if len(full[0]) != len(FullColumns) {
		t.Errorf("want %d full cols got %d", len(FullColumns), len(full[0]))
	}
	compact := parseOptionChain(raw, "NIFTY", "25-Jan-2024", "compact")
	if len(compact) != 1 || len(compact[0]) != len(CompactColumns) {
		t.Errorf("compact filter failed: %v", compact)
	}
}

func TestIsIndexSymbol(t *testing.T) {
	if !isIndexSymbol("BANKNIFTY") || !isIndexSymbol("NIFTY") {
		t.Error("index not detected")
	}
	if isIndexSymbol("SBIN") {
		t.Error("SBIN flagged as index")
	}
}

func TestNormalizeFOGrowth(t *testing.T) {
	fy, ty, err := NormalizeBusinessGrowthFinancialYear("2025-2026", "")
	if err != nil || fy != "2025" || ty != "2026" {
		t.Errorf("got %s %s %v", fy, ty, err)
	}
	m, y, err := NormalizeBusinessGrowthDailyArgs("Mar-26", "")
	if err != nil || m != "Mar" || y != "2026" {
		t.Errorf("FO daily should expand to 2026, got %s %s %v", m, y, err)
	}
}

func TestBusinessGrowthDataFrame(t *testing.T) {
	raw := map[string]interface{}{"data": []interface{}{
		map[string]interface{}{"data": map[string]interface{}{"a": "1,000", "date": "x"}},
	}}
	df := businessGrowthDataFrame(raw)
	if df[0]["a"] != 1000.0 || df[0]["date"] != "x" {
		t.Errorf("bad frame: %v", df[0])
	}
}

func TestParseBanLines(t *testing.T) {
	got := parseBanLines([]string{"col1,col2", "1,SBIN", "2,INFY"})
	if len(got) != 2 || got[0] != "SBIN" || got[1] != "INFY" {
		t.Errorf("bad parse: %v", got)
	}
}
