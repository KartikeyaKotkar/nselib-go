package nsdlfpi

import (
	"testing"

	"github.com/KartikeyaKotkar/nselib-go"
)

func TestExtractHiddenFields(t *testing.T) {
	html := `<input name="__VIEWSTATE" id="__VIEWSTATE" value="abc" />
<input name="__VIEWSTATEGENERATOR" id="__VIEWSTATEGENERATOR" value="def" />
<input name="__EVENTVALIDATION" id="__EVENTVALIDATION" value="ghi" />`
	f := ExtractHiddenFields(html)
	if f["__VIEWSTATE"] != "abc" || f["__EVENTVALIDATION"] != "ghi" {
		t.Errorf("bad fields: %v", f)
	}
}

func TestClassifyTable(t *testing.T) {
	if ClassifyTable([]string{"X", "Investment Route"}) != "investment" {
		t.Error("investment not classified")
	}
	if ClassifyTable([]string{"Derivative Products"}) != "derivative" {
		t.Error("derivative not classified")
	}
	if ClassifyTable([]string{"Other"}) != "" {
		t.Error("other should be empty")
	}
}

func TestParseNumeric(t *testing.T) {
	cases := []struct {
		in   interface{}
		want float64
		ok   bool
	}{
		{"1,234.5", 1234.5, true}, {"(100)", -100, true},
		{"Rs. 50", 50, true}, {"-", 0, false}, {"nan", 0, false},
		{12.5, 12.5, true}, {nil, 0, false},
	}
	for _, c := range cases {
		got, ok := ParseNumeric(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("%v: got %v %v", c.in, got, ok)
		}
	}
}

func TestCoerceTradeDate(t *testing.T) {
	d, err := CoerceTradeDate("30-10-2025")
	if err != nil || d.Format("2006-01-02") != "2025-10-30" {
		t.Errorf("got %v %v", d, err)
	}
	if _, err := CoerceTradeDate("blah"); err == nil {
		t.Error("want error")
	}
}

func investmentFixture() nselib.HTMLTable {
	return nselib.HTMLTable{
		Headers: []string{"Reporting Date", "Debt / Equity", "Investment Route",
			"Gross Purchases", "Gross Sales", "Net Investment", "Net Investment US($) Mn", "Conversion"},
		Rows: [][]string{
			{"30-Oct-2025", "Equity", "Stock Exchange", "100", "60", "40", "5", "83.5"},
			{"30-Oct-2025", "Note", "x", "1", "1", "0", "0", ""},
		},
	}
}

func TestParseInvestmentTable(t *testing.T) {
	df := ParseInvestmentTable(investmentFixture())
	if len(df) != 1 {
		t.Fatalf("want 1 row (note dropped, conversion required) got %d: %v", len(df), df)
	}
	if df[0]["ASSET_CLASS"] != "Equity" || df[0]["GROSS_PURCHASES_RS_CR"] != 100.0 {
		t.Errorf("bad row: %v", df[0])
	}
}

func TestParseDerivativeTable(t *testing.T) {
	tb := nselib.HTMLTable{
		Headers: []string{"Reporting Date", "Derivative Products",
			"Buy No. of Contracts", "Buy Amount in Crore",
			"Sell No. of Contracts", "Sell Amount in Crore",
			"Open Interest No. of Contracts", "Open Interest Amount in Crore"},
		Rows: [][]string{
			{"30-Oct-2025", "index futures", "10", "1", "5", "0.5", "20", "2"},
			{"30-Oct-2025", "Weird Product", "1", "1", "1", "1", "1", "1"},
		},
	}
	df := ParseDerivativeTable(tb)
	if len(df) != 1 {
		t.Fatalf("want 1 row (allow-list) got %v", df)
	}
	if df[0]["DERIVATIVE_PRODUCT"] != "Index Futures" {
		t.Errorf("title case failed: %v", df[0])
	}
}

func TestParseReportBundle(t *testing.T) {
	html := `<html><body>
<table><tr><th>Reporting Date</th><th>Investment Route</th><th>Conversion</th></tr>
<tr><td>30-Oct-2025</td><td>Route</td><td>83</td></tr></table>
<table><tr><th>Derivative Products</th></tr><tr><td>Index Futures</td></tr></table>
</body></html>`
	b := ParseReportBundle(html, "test", nil)
	if b.SourcePage != "test" {
		t.Error("source lost")
	}
	if len(b.Investment) != 0 {
		t.Errorf("investment should drop (no asset class): %v", b.Investment)
	}
}
