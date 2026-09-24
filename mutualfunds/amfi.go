// Package mutualfunds ports nselib/mutual_funds to Go.
package mutualfunds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
	"github.com/extrame/xls"
	"github.com/xuri/excelize/v2"
)

// AMFIMonthlyPageURL is the AMFI monthly archive page.
const AMFIMonthlyPageURL = "https://www.amfiindia.com/research-information/amfi-monthly"

var reportLinkPattern = regexp.MustCompile(
	`/spages/am([a-z]+)(\d{4})repo(?:revised)?\.(pdf|xls|xlsx|htm|html)$`)

var monthMap = map[string]int{
	"jan": 1, "january": 1, "feb": 2, "february": 2, "mar": 3, "march": 3,
	"apr": 4, "april": 4, "may": 5, "jun": 6, "june": 6, "jul": 7, "july": 7,
	"aug": 8, "august": 8, "sep": 9, "september": 9, "oct": 10, "october": 10,
	"nov": 11, "november": 11, "dec": 12, "december": 12,
}

// DefaultFileTypePriority mirrors Python's default.
var DefaultFileTypePriority = []string{"pdf", "xls", "xlsx", "html", "htm"}

// OutputColumns mirrors Python's _OUTPUT_COLUMNS.
var OutputColumns = []string{
	"REPORT_MONTH", "PERIOD_LABEL", "SOURCE_URL", "SOURCE_FILE_TYPE",
	"SECTION_NAME", "TABLE_INDEX", "ROW_INDEX", "ROW_JSON", "ROW_TEXT",
}

var hrefPattern = regexp.MustCompile(`href=["']([^"']+)["']`)

// LinkRecord mirrors one row of the AMFI links frame.
type LinkRecord struct {
	ReportMonth    time.Time
	PeriodLabel    string
	ReportYear     int
	ReportMonthNum int
	FileType       string
	IsRevised      bool
	SourceURL      string
}

var amfiHTTP = &http.Client{Timeout: 90 * time.Second}

// ExtractLinkRecords scrapes AMFI report links from archive HTML.
func ExtractLinkRecords(pageHTML string) []LinkRecord {
	base, _ := url.Parse(AMFIMonthlyPageURL)
	seen := map[string]bool{}
	var out []LinkRecord
	for _, m := range hrefPattern.FindAllStringSubmatch(pageHTML, -1) {
		href := strings.TrimSpace(m[1])
		u, err := url.Parse(href)
		if err != nil {
			continue
		}
		full := base.ResolveReference(u).String()
		lower := strings.ToLower(full)
		match := reportLinkPattern.FindStringSubmatch(lower)
		if match == nil {
			continue
		}
		monthVal, ok := monthMap[strings.ToLower(match[1])]
		if !ok {
			continue
		}
		var yearVal int
		fmt.Sscanf(match[2], "%d", &yearVal)
		fileType := strings.ToLower(match[3])
		key := full + "|" + fileType
		if seen[key] {
			continue
		}
		seen[key] = true
		rm := time.Date(yearVal, time.Month(monthVal), 1, 0, 0, 0, 0, time.UTC)
		out = append(out, LinkRecord{
			ReportMonth:    rm,
			PeriodLabel:    rm.Format("Jan-2006"),
			ReportYear:     yearVal,
			ReportMonthNum: monthVal,
			FileType:       fileType,
			IsRevised:      strings.Contains(lower, "reporevised."),
			SourceURL:      full,
		})
	}
	return out
}

// PriorityRank ranks a file type within a priority list.
func PriorityRank(fileType string, priority []string) int {
	n := strings.ToLower(strings.TrimSpace(fileType))
	for i, p := range priority {
		if p == n {
			return i
		}
	}
	return len(priority)
}

// PreferredLinks keeps the best-priority link per report month.
func PreferredLinks(links []LinkRecord, priority []string) []LinkRecord {
	cp := append([]LinkRecord{}, links...)
	sort.Slice(cp, func(i, j int) bool {
		if cp[i].ReportMonth.Equal(cp[j].ReportMonth) {
			ri, rj := PriorityRank(cp[i].FileType, priority), PriorityRank(cp[j].FileType, priority)
			if ri != rj {
				return ri < rj
			}
			if cp[i].IsRevised != cp[j].IsRevised {
				return cp[j].IsRevised
			}
			return cp[i].SourceURL < cp[j].SourceURL
		}
		return cp[i].ReportMonth.Before(cp[j].ReportMonth)
	})
	var out []LinkRecord
	seen := map[string]bool{}
	for _, l := range cp {
		k := l.ReportMonth.Format("2006-01")
		if !seen[k] {
			seen[k] = true
			out = append(out, l)
		}
	}
	return out
}

// NormalizeReportMonth parses a month value and returns the first of that month.
// Accepts dd-mm-YYYY, dd-Mon-YYYY, Mon-YYYY, YYYY-mm-dd, YYYY-mm.
func NormalizeReportMonth(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	layouts := []string{
		"02-01-2006", "2-01-2006", "02-Jan-2006", "2-Jan-2006",
		"Jan-2006", "January-2006", "2006-01-02", "2006-01", "01-2006",
	}
	var t time.Time
	var err error
	for _, l := range layouts {
		if t, err = time.Parse(l, value); err == nil {
			return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
		}
	}
	// last resort: month name + year tokens
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == ' ' || r == '-' || r == '/' })
	if len(parts) == 2 {
		if m, ok := monthMap[strings.ToLower(parts[0])]; ok {
			var y int
			if _, err := fmt.Sscanf(parts[1], "%d", &y); err == nil {
				if y < 100 {
					y += 2000
				}
				return time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC), nil
			}
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse report month: %s", value)
}

// CoerceText trims cell text, returning "" for blanks/NaN.
func CoerceText(v interface{}) string {
	if v == nil {
		return ""
	}
	s := strings.ReplaceAll(fmt.Sprintf("%v", v), "\u00a0", " ")
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "nan") || strings.EqualFold(s, "none") {
		return ""
	}
	return s
}

// RowPayloadFromValues builds C1..Cn payload skipping blank cells.
func RowPayloadFromValues(values []string) map[string]string {
	payload := map[string]string{}
	for i, v := range values {
		if t := CoerceText(v); t != "" {
			payload[fmt.Sprintf("C%d", i+1)] = t
		}
	}
	return payload
}

// BuildOutputRow creates one normalized output record; nil when payload empty.
func BuildOutputRow(reportMonth time.Time, periodLabel, sourceURL, sourceFileType, section string, tableIndex, rowIndex int, payload map[string]string) nselib.Record {
	if len(payload) == 0 {
		return nil
	}
	vals := make([]string, 0, len(payload))
	for i := 1; ; i++ {
		v, ok := payload[fmt.Sprintf("C%d", i)]
		if !ok {
			break
		}
		vals = append(vals, v)
	}
	// include any non-sequential keys deterministically
	for k, v := range payload {
		known := false
		for i := 1; i <= len(vals); i++ {
			if k == fmt.Sprintf("C%d", i) {
				known = true
				break
			}
		}
		if !known {
			vals = append(vals, v)
		}
	}
	js, _ := json.Marshal(payload)
	return nselib.Record{
		"REPORT_MONTH":     reportMonth.Format("2006-01-02"),
		"PERIOD_LABEL":     periodLabel,
		"SOURCE_URL":       sourceURL,
		"SOURCE_FILE_TYPE": sourceFileType,
		"SECTION_NAME":     section,
		"TABLE_INDEX":      tableIndex,
		"ROW_INDEX":        rowIndex,
		"ROW_JSON":         string(js),
		"ROW_TEXT":         strings.Join(vals, " | "),
	}
}

// ParseExcelReport parses xls (extrame/xls) and xlsx (excelize) into output rows.
func ParseExcelReport(content []byte, reportMonth time.Time, periodLabel, sourceURL, fileType string) (nselib.DataFrame, error) {
	var sheets []struct {
		name string
		rows [][]string
	}
	if fileType == "xls" {
		wb, err := xls.OpenReader(bytes.NewReader(content), "utf-8")
		if err != nil {
			return nil, nselib.NewDataNotFoundError("open AMFI xls: " + err.Error())
		}
		for i := 0; i < wb.NumSheets(); i++ {
			sh := wb.GetSheet(i)
			if sh == nil {
				continue
			}
			var rows [][]string
			for r := 0; r <= int(sh.MaxRow); r++ {
				row := sh.Row(r)
				rec := make([]string, int(row.LastCol()))
				for c := range rec {
					rec[c] = row.Col(c)
				}
				rows = append(rows, rec)
			}
			sheets = append(sheets, struct {
				name string
				rows [][]string
			}{sh.Name, rows})
		}
	} else {
		f, err := excelize.OpenReader(bytes.NewReader(content))
		if err != nil {
			return nil, nselib.NewDataNotFoundError("open AMFI xlsx: " + err.Error())
		}
		for _, name := range f.GetSheetList() {
			rows, err := f.GetRows(name)
			if err != nil {
				continue
			}
			sheets = append(sheets, struct {
				name string
				rows [][]string
			}{name, rows})
		}
	}
	var out nselib.DataFrame
	for _, sh := range sheets {
		rowIdx := 0
		for _, r := range sh.rows {
			rowIdx++
			if row := BuildOutputRow(reportMonth, periodLabel, sourceURL, fileType,
				"Sheet:"+sh.name, 1, rowIdx, RowPayloadFromValues(r)); row != nil {
				out = append(out, row)
			}
		}
	}
	if out == nil {
		out = nselib.DataFrame{}
	}
	return out, nil
}

// ParseHTMLReport parses HTML tables into output rows.
func ParseHTMLReport(content []byte, reportMonth time.Time, periodLabel, sourceURL, fileType string) (nselib.DataFrame, error) {
	tables := nselib.ExtractHTMLTables(string(content))
	var out nselib.DataFrame
	for ti, tb := range tables {
		all := append([][]string{tb.Headers}, tb.Rows...)
		for ri, r := range all {
			if row := BuildOutputRow(reportMonth, periodLabel, sourceURL, fileType,
				fmt.Sprintf("Table:%d", ti+1), ti+1, ri+1, RowPayloadFromValues(r)); row != nil {
				out = append(out, row)
			}
		}
	}
	if out == nil {
		out = nselib.DataFrame{}
	}
	return out, nil
}

// ParsePDFReport extracts text lines per page (text-only; mirrors pypdf fallback).
// NOTE: pdfplumber table structure is not available in pure-Go MIT libraries;
// PDF output matches the line-based shape of Python's pypdf fallback path.
func ParsePDFReport(extract func() ([]string, error), reportMonth time.Time, periodLabel, sourceURL string) (nselib.DataFrame, error) {
	pages, err := extract()
	if err != nil {
		return nil, nselib.NewDataNotFoundError("Unable to parse AMFI PDF report: " + sourceURL + " :: " + err.Error())
	}
	var out nselib.DataFrame
	for pi, text := range pages {
		for li, line := range strings.Split(text, "\n") {
			if row := BuildOutputRow(reportMonth, periodLabel, sourceURL, "pdf",
				fmt.Sprintf("Page:%d", pi+1), (pi+1)*1000, li+1, RowPayloadFromValues([]string{line})); row != nil {
				out = append(out, row)
			}
		}
	}
	if out == nil {
		out = nselib.DataFrame{}
	}
	return out, nil
}

// ParseReportContent dispatches on file type.
func ParseReportContent(content []byte, reportMonth time.Time, periodLabel, sourceURL, fileType, pdfText string, pdfErr error) (nselib.DataFrame, error) {
	switch fileType {
	case "xls", "xlsx":
		return ParseExcelReport(content, reportMonth, periodLabel, sourceURL, fileType)
	case "html", "htm":
		return ParseHTMLReport(content, reportMonth, periodLabel, sourceURL, fileType)
	case "pdf":
		if pdfErr != nil {
			return nil, nselib.NewDataNotFoundError("Unable to parse AMFI PDF report: " + sourceURL + " :: " + pdfErr.Error())
		}
		pages := strings.Split(pdfText, "\n\x0c\n")
		return ParsePDFReport(func() ([]string, error) { return pages, nil },
			reportMonth, periodLabel, sourceURL)
	}
	return nil, nselib.NewDataNotFoundError("Unsupported AMFI monthly report type: " + fileType)
}

// downloadReport fetches report bytes with browser UA.
func downloadReport(sourceURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", sourceURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")
	resp, err := amfiHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, nselib.NewDataNotFoundError(fmt.Sprintf("AMFI report URL not reachable: %s (status=%d)", sourceURL, resp.StatusCode))
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// AMFIMonthlyReportLinks lists all report links on the AMFI archive page.
func AMFIMonthlyReportLinks() (nselib.DataFrame, error) {
	req, _ := http.NewRequest("GET", AMFIMonthlyPageURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")
	resp, err := amfiHTTP.Do(req)
	if err != nil {
		return nil, nselib.NewDataNotFoundError("Unable to access AMFI monthly archive page: " + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, nselib.NewDataNotFoundError(fmt.Sprintf("Unable to access AMFI monthly archive page (status=%d)", resp.StatusCode))
	}
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	links := ExtractLinkRecords(buf.String())
	out := make(nselib.DataFrame, 0, len(links))
	for _, l := range links {
		out = append(out, nselib.Record{
			"REPORT_MONTH":     l.ReportMonth.Format("2006-01-02"),
			"PERIOD_LABEL":     l.PeriodLabel,
			"REPORT_YEAR":      l.ReportYear,
			"REPORT_MONTH_NUM": l.ReportMonthNum,
			"FILE_TYPE":        l.FileType,
			"IS_REVISED":       l.IsRevised,
			"SOURCE_URL":       l.SourceURL,
		})
	}
	return out, nil
}

// AMFIMonthlyData fetches a single month report as normalized rows.
func AMFIMonthlyData(reportMonth string, fileTypePriority ...string) (nselib.DataFrame, error) {
	norm, err := NormalizeReportMonth(reportMonth)
	if err != nil {
		return nil, err
	}
	priority := DefaultFileTypePriority
	if len(fileTypePriority) > 0 {
		priority = fileTypePriority
	}
	links, err := AMFIMonthlyReportLinksRecords()
	if err != nil {
		return nil, err
	}
	var monthLinks []LinkRecord
	for _, l := range links {
		if l.ReportMonth.Equal(norm) {
			monthLinks = append(monthLinks, l)
		}
	}
	if len(monthLinks) == 0 {
		return nselib.DataFrame{}, nil
	}
	sel := PreferredLinks(monthLinks, priority)[0]
	content, err := downloadReport(sel.SourceURL)
	if err != nil {
		return nil, err
	}
	var pdfText string
	var pdfErr error
	if sel.FileType == "pdf" {
		pdfText, pdfErr = extractPDFText(content)
	}
	return ParseReportContent(content, norm, sel.PeriodLabel, sel.SourceURL, sel.FileType, pdfText, pdfErr)
}

// AMFIMonthlyReportLinksRecords returns parsed link records.
func AMFIMonthlyReportLinksRecords() ([]LinkRecord, error) {
	df, err := AMFIMonthlyReportLinks()
	if err != nil {
		return nil, err
	}
	var out []LinkRecord
	for _, r := range df {
		rm, _ := time.Parse("2006-01-02", r["REPORT_MONTH"].(string))
		out = append(out, LinkRecord{
			ReportMonth:    rm,
			PeriodLabel:    r["PERIOD_LABEL"].(string),
			ReportYear:     r["REPORT_YEAR"].(int),
			ReportMonthNum: r["REPORT_MONTH_NUM"].(int),
			FileType:       r["FILE_TYPE"].(string),
			IsRevised:      r["IS_REVISED"].(bool),
			SourceURL:      r["SOURCE_URL"].(string),
		})
	}
	return out, nil
}

// AMFIMonthlyHistoricalData fetches reports across a month range.
func AMFIMonthlyHistoricalData(fromMonth, toMonth string, fileTypePriority []string, includeAllVariants, strict bool) (nselib.DataFrame, error) {
	priority := DefaultFileTypePriority
	if len(fileTypePriority) > 0 {
		priority = fileTypePriority
	}
	links, err := AMFIMonthlyReportLinksRecords()
	if err != nil {
		return nil, err
	}
	if fromMonth != "" {
		anchor, err := NormalizeReportMonth(fromMonth)
		if err != nil {
			return nil, err
		}
		var keep []LinkRecord
		for _, l := range links {
			if !l.ReportMonth.Before(anchor) {
				keep = append(keep, l)
			}
		}
		links = keep
	}
	if toMonth != "" {
		anchor, err := NormalizeReportMonth(toMonth)
		if err != nil {
			return nil, err
		}
		var keep []LinkRecord
		for _, l := range links {
			if !l.ReportMonth.After(anchor) {
				keep = append(keep, l)
			}
		}
		links = keep
	}
	if len(links) == 0 {
		return nselib.DataFrame{}, nil
	}
	var selected []LinkRecord
	if includeAllVariants {
		selected = append([]LinkRecord{}, links...)
		sort.Slice(selected, func(i, j int) bool {
			if selected[i].ReportMonth.Equal(selected[j].ReportMonth) {
				if selected[i].FileType != selected[j].FileType {
					return selected[i].FileType < selected[j].FileType
				}
				return selected[i].SourceURL < selected[j].SourceURL
			}
			return selected[i].ReportMonth.Before(selected[j].ReportMonth)
		})
	} else {
		selected = PreferredLinks(links, priority)
	}
	seen := map[string]bool{}
	var months []time.Time
	for _, l := range selected {
		k := l.ReportMonth.Format("2006-01")
		if !seen[k] {
			seen[k] = true
			months = append(months, l.ReportMonth)
		}
	}
	var out nselib.DataFrame
	for _, m := range months {
		frame, err := AMFIMonthlyData(m.Format("Jan-2006"), priority...)
		if err != nil {
			if strict {
				return nil, err
			}
			continue
		}
		out = append(out, frame...)
	}
	if out == nil {
		out = nselib.DataFrame{}
	}
	return out, nil
}
