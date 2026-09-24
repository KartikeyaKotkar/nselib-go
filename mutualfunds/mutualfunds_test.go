package mutualfunds

import (
	"testing"
	"time"
)

func TestExtractLinkRecords(t *testing.T) {
	html := `<a href="/spages/amjan2025repo.pdf">Jan</a>
<a href="/spages/amfeb2025reporevised.xls">Feb</a>
<a href="/other/page.html">Nope</a>`
	links := ExtractLinkRecords(html)
	if len(links) != 2 {
		t.Fatalf("want 2 links got %d", len(links))
	}
	if links[0].FileType != "pdf" || links[0].PeriodLabel != "Jan-2025" {
		t.Errorf("bad link: %+v", links[0])
	}
	if !links[1].IsRevised || links[1].FileType != "xls" {
		t.Errorf("bad revised link: %+v", links[1])
	}
}

func TestPreferredLinks(t *testing.T) {
	m := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	links := []LinkRecord{
		{ReportMonth: m, FileType: "xls", SourceURL: "b"},
		{ReportMonth: m, FileType: "pdf", SourceURL: "a"},
	}
	sel := PreferredLinks(links, DefaultFileTypePriority)
	if len(sel) != 1 || sel[0].FileType != "pdf" {
		t.Errorf("want pdf got %v", sel)
	}
}

func TestNormalizeReportMonth(t *testing.T) {
	for in, want := range map[string]string{
		"Jan-2025": "2025-01-01", "15-02-2025": "2025-02-01",
		"2025-03-20": "2025-03-01", "March-2025": "2025-03-01",
	} {
		got, err := NormalizeReportMonth(in)
		if err != nil || got.Format("2006-01-02") != want {
			t.Errorf("%s: got %v %v", in, got, err)
		}
	}
	if _, err := NormalizeReportMonth("blah"); err == nil {
		t.Error("want error on garbage")
	}
}

func TestRowPayload(t *testing.T) {
	p := RowPayloadFromValues([]string{" a ", "", "nan", "1,000"})
	if len(p) != 2 || p["C1"] != "a" || p["C4"] != "1,000" {
		t.Errorf("bad payload: %v", p)
	}
	m := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	row := BuildOutputRow(m, "Jan-2025", "u", "pdf", "Page:1", 1000, 1, p)
	if row["ROW_TEXT"] != "a | 1,000" {
		t.Errorf("bad text: %v", row["ROW_TEXT"])
	}
	if BuildOutputRow(m, "x", "u", "pdf", "s", 1, 1, map[string]string{}) != nil {
		t.Error("empty payload should be nil")
	}
}

func TestParseHTMLReport(t *testing.T) {
	html := `<table><tr><th>H1</th></tr><tr><td>v1</td></tr></table>`
	m := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	df, err := ParseHTMLReport([]byte(html), m, "Jan-2025", "u", "html")
	if err != nil {
		t.Fatal(err)
	}
	if len(df) != 2 || df[0]["SECTION_NAME"] != "Table:1" {
		t.Errorf("bad rows: %v", df)
	}
}

func TestParsePDFReportLines(t *testing.T) {
	m := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	df, err := ParsePDFReport(func() ([]string, error) {
		return []string{"line1\n\nline2"}, nil
	}, m, "Jan-2025", "u")
	if err != nil {
		t.Fatal(err)
	}
	if len(df) != 2 {
		t.Errorf("want 2 lines got %v", df)
	}
}

func TestParseReportDispatch(t *testing.T) {
	m := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := ParseReportContent([]byte("x"), m, "l", "u", "docx", "", nil); err == nil {
		t.Error("want error on unsupported type")
	}
}
