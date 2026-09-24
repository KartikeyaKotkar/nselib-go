package nsdlfpi

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

var hiddenFieldPattern = map[string]*regexp.Regexp{}

// ExtractHiddenFields pulls ASP.NET hidden fields via regex like Python.
func ExtractHiddenFields(htmlText string) map[string]string {
	out := map[string]string{}
	for _, name := range ArchiveHiddenFields {
		re := regexp.MustCompile(`name="` + name + `"\s+id="` + name + `"\s+value="([^"]*)"`)
		hiddenFieldPattern[name] = re
		if m := re.FindStringSubmatch(htmlText); m != nil {
			out[name] = m[1]
		} else {
			out[name] = ""
		}
	}
	return out
}

// ClassifyTable returns "investment", "derivative" or "".
func ClassifyTable(headers []string) string {
	for _, h := range headers {
		l := strings.ToLower(h)
		if strings.Contains(l, "investment route") {
			return "investment"
		}
		if strings.Contains(l, "derivative products") {
			return "derivative"
		}
	}
	return ""
}

// NormalizeText trims text, "" for blanks/NaN.
func NormalizeText(v interface{}) string {
	if v == nil {
		return ""
	}
	var s string
	switch t := v.(type) {
	case string:
		s = t
	case float64:
		s = strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		s = strconv.Itoa(t)
	default:
		s = ""
	}
	s = strings.TrimSpace(strings.ReplaceAll(s, "\u00a0", " "))
	if s == "" || strings.EqualFold(s, "nan") {
		return ""
	}
	return s
}

// ParseNumeric mirrors Python's _parse_numeric (parens negative, commas, Rs).
func ParseNumeric(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}
	if f, ok := v.(float64); ok {
		return f, true
	}
	s := strings.TrimSpace(NormalizeText(v))
	if s == "" || strings.EqualFold(s, "nan") || s == "none" || s == "-" || s == "--" {
		return 0, false
	}
	neg := strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")")
	s = strings.Trim(s, "()")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "Rs.", "")
	s = strings.ReplaceAll(s, "Rs", "")
	s = strings.TrimSpace(s)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	if neg {
		f = -f
	}
	return f, true
}

// CoerceTradeDate accepts dd-mm-YYYY and dd-Mon-YYYY strings.
func CoerceTradeDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	for _, l := range []string{nselib.LayoutDDMMYYYY, ReportDateLayout} {
		if t, err := time.Parse(l, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("Invalid NSDL trade date: %s", value)
}

// titleCase mirrors Python str.title for product names.
func titleCase(s string) string {
	words := strings.Fields(strings.ToLower(s))
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// ParseInvestmentTable normalizes one investment HTML table.
func ParseInvestmentTable(tb nselib.HTMLTable) nselib.DataFrame {
	renamed := renameColumns(append([]string{}, tb.Headers...), investmentRename)
	avail := intersect(renamed, InvestmentColumns)
	if len(avail) == 0 {
		return nselib.DataFrame{}
	}
	var out nselib.DataFrame
	seen := map[string]bool{}
	lastDate := ""
	for _, row := range tb.Rows {
		rec := zipRow(renamed, row)
		date := NormalizeText(rec["REPORT_DATE"])
		if d, err := time.Parse(ReportDateLayout, date); err == nil {
			lastDate = d.Format(ReportDateLayout)
		} else if date == "" {
			date = lastDate // ffill
		} else {
			continue
		}
		if date == "" {
			continue
		}
		rec["REPORT_DATE"] = date
		rec["ASSET_CLASS"] = NormalizeText(rec["ASSET_CLASS"])
		rec["INVESTMENT_ROUTE"] = NormalizeText(rec["INVESTMENT_ROUTE"])
		asset, _ := rec["ASSET_CLASS"].(string)
		if asset == "" || strings.EqualFold(asset, "note") {
			continue
		}
		for _, c := range []string{"GROSS_PURCHASES_RS_CR", "GROSS_SALES_RS_CR",
			"NET_INVESTMENT_RS_CR", "NET_INVESTMENT_USD_MN", "USD_INR_CONVERSION"} {
			if f, ok := ParseNumeric(rec[c]); ok {
				rec[c] = f
			} else {
				rec[c] = nil
			}
		}
		if rec["USD_INR_CONVERSION"] == nil {
			continue
		}
		key := date + "|" + rec["ASSET_CLASS"].(string) + "|" + NormalizeText(rec["INVESTMENT_ROUTE"])
		if seen[key] {
			continue
		}
		seen[key] = true
		full := nselib.Record{}
		for _, c := range InvestmentColumns {
			full[c] = rec[c]
		}
		out = append(out, full)
	}
	if out == nil {
		out = nselib.DataFrame{}
	}
	return out
}

// ParseDerivativeTable normalizes one derivative HTML table.
func ParseDerivativeTable(tb nselib.HTMLTable) nselib.DataFrame {
	renamed := renameColumns(append([]string{}, tb.Headers...), derivativeRename)
	avail := intersect(renamed, DerivativeColumns)
	if len(avail) == 0 {
		return nselib.DataFrame{}
	}
	var out nselib.DataFrame
	seen := map[string]bool{}
	lastDate := ""
	for _, row := range tb.Rows {
		rec := zipRow(renamed, row)
		date := NormalizeText(rec["REPORT_DATE"])
		if d, err := time.Parse(ReportDateLayout, date); err == nil {
			lastDate = d.Format(ReportDateLayout)
		} else if date == "" {
			date = lastDate
		} else {
			continue
		}
		if date == "" {
			continue
		}
		rec["REPORT_DATE"] = date
		prod := titleCase(strings.Join(strings.Fields(NormalizeText(rec["DERIVATIVE_PRODUCT"])), " "))
		if !DailyDerivativeProducts[prod] {
			continue
		}
		rec["DERIVATIVE_PRODUCT"] = prod
		for _, c := range []string{"BUY_CONTRACTS", "BUY_AMOUNT_CR", "SELL_CONTRACTS",
			"SELL_AMOUNT_CR", "OPEN_INTEREST_CONTRACTS", "OPEN_INTEREST_AMOUNT_CR"} {
			if f, ok := ParseNumeric(rec[c]); ok {
				rec[c] = f
			} else {
				rec[c] = nil
			}
		}
		key := date + "|" + prod
		if seen[key] {
			continue
		}
		seen[key] = true
		full := nselib.Record{}
		for _, c := range DerivativeColumns {
			full[c] = rec[c]
		}
		out = append(out, full)
	}
	if out == nil {
		out = nselib.DataFrame{}
	}
	return out
}

func investmentRename(h string) string {
	l := strings.ToLower(h)
	switch {
	case strings.Contains(l, "reporting date"):
		return "REPORT_DATE"
	case strings.Contains(l, "investment route"):
		return "INVESTMENT_ROUTE"
	case strings.Contains(l, "gross purchases"):
		return "GROSS_PURCHASES_RS_CR"
	case strings.Contains(l, "gross sales"):
		return "GROSS_SALES_RS_CR"
	case strings.Contains(l, "net investment") && !strings.Contains(l, "us($"):
		return "NET_INVESTMENT_RS_CR"
	case strings.Contains(l, "net investment us($"):
		return "NET_INVESTMENT_USD_MN"
	case strings.Contains(l, "conversion"):
		return "USD_INR_CONVERSION"
	case strings.Contains(l, "debt") && strings.Contains(l, "equity"):
		return "ASSET_CLASS"
	}
	return h
}

func derivativeRename(h string) string {
	l := strings.ToLower(h)
	switch {
	case strings.Contains(l, "reporting date"):
		return "REPORT_DATE"
	case strings.Contains(l, "derivative products"):
		return "DERIVATIVE_PRODUCT"
	case strings.Contains(l, "buy") && strings.Contains(l, "no. of contracts"):
		return "BUY_CONTRACTS"
	case strings.Contains(l, "buy") && strings.Contains(l, "amount in crore"):
		return "BUY_AMOUNT_CR"
	case strings.Contains(l, "sell") && strings.Contains(l, "no. of contracts"):
		return "SELL_CONTRACTS"
	case strings.Contains(l, "sell") && strings.Contains(l, "amount in crore"):
		return "SELL_AMOUNT_CR"
	case strings.Contains(l, "open interest") && strings.Contains(l, "no. of contracts"):
		return "OPEN_INTEREST_CONTRACTS"
	case strings.Contains(l, "open interest") && strings.Contains(l, "amount in crore"):
		return "OPEN_INTEREST_AMOUNT_CR"
	}
	return h
}

func renameColumns(headers []string, fn func(string) string) []string {
	out := make([]string, len(headers))
	for i, h := range headers {
		out[i] = fn(strings.TrimSpace(h))
	}
	return out
}

func zipRow(headers []string, row []string) nselib.Record {
	rec := nselib.Record{}
	for i, h := range headers {
		if i < len(row) {
			rec[h] = row[i]
		}
	}
	return rec
}

func intersect(headers, want []string) []string {
	set := map[string]bool{}
	for _, h := range headers {
		set[h] = true
	}
	var out []string
	for _, w := range want {
		if set[w] {
			out = append(out, w)
		}
	}
	return out
}

// ParseReportBundle classifies tables and builds the bundle.
func ParseReportBundle(htmlText, sourcePage string, asOfDate *time.Time) *ReportBundle {
	b := &ReportBundle{
		Investment: nselib.DataFrame{},
		Derivative: nselib.DataFrame{},
		SourcePage: sourcePage,
		AsOfDate:   asOfDate,
	}
	for _, tb := range nselib.ExtractHTMLTables(htmlText) {
		switch ClassifyTable(tb.Headers) {
		case "investment":
			b.Investment = ParseInvestmentTable(tb)
		case "derivative":
			b.Derivative = ParseDerivativeTable(tb)
		}
	}
	return b
}
