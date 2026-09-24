package derivatives

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
	"github.com/extrame/xls"
)

// DailyVolatility fetches F&O daily volatility with 3-URL fallback.
func DailyVolatility(tradeDate string) (nselib.DataFrame, error) {
	t, err := time.Parse(nselib.LayoutDDMMYYYY, tradeDate)
	if err != nil {
		return nil, err
	}
	payload := fmt.Sprintf("FOVOLT_%s.csv", t.Format(nselib.LayoutDDMMYYYYCompact))
	urls := []string{
		"https://nsearchives.nseindia.com/archives/nsccl/volt/" + payload,
		"https://archives.nseindia.com/archives/nsccl/volt/" + payload,
		"https://www.nseindia.com/api/reports?archives=" +
			"%5B%7B%22name%22%3A%22F%26O%20-%20Daily%20Volatility%22%2C%22type%22%3A%22archives%22%2C%22category%22" +
			fmt.Sprintf("%%3A%%22derivatives%%22%%2C%%22section%%22%%3A%%22equity%%22%%7D%%5D&date=%s",
				t.Format(nselib.LayoutDDMMMYYYY)) + "&type=equity&mode=single",
	}
	var lastErr error
	for _, u := range urls {
		body, err := defaultClient.FetchBytes(u, "https://www.nseindia.com/all-reports-derivatives")
		if err != nil {
			lastErr = err
			continue
		}
		df, err := nselib.ParseCSV(bytes.NewReader(body))
		if err != nil {
			lastErr = err
			continue
		}
		out := make(nselib.DataFrame, 0, len(df))
		for _, r := range df {
			empty := true
			for _, v := range r {
				if s, _ := v.(string); s != "" {
					empty = false
					break
				}
			}
			if !empty {
				out = append(out, r)
			}
		}
		return out, nil
	}
	if lastErr == nil {
		lastErr = nselib.NewDataNotFoundError("no FO daily volatility data for " + tradeDate)
	}
	return nil, lastErr
}

// CategoryTurnoverFO fetches F&O category turnover (.xls, single header block).
func CategoryTurnoverFO(tradeDate string) (nselib.DataFrame, error) {
	t, err := time.Parse(nselib.LayoutDDMMYYYY, tradeDate)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://archives.nseindia.com/archives/fo/cat/fo_cat_turnover_%s.xls",
		t.Format(nselib.LayoutDDMMYYCompact))
	body, err := defaultClient.FetchBytes(url, "")
	if err != nil {
		return nil, nselib.NewDataNotFoundError(fmt.Sprintf("No data available for : %s", tradeDate))
	}
	return parseCategoryTurnoverXLS(body)
}

// parseCategoryTurnoverXLS replicates the Python single-block xlrd scan.
func parseCategoryTurnoverXLS(body []byte) (nselib.DataFrame, error) {
	wb, err := xls.OpenReader(bytes.NewReader(body), "utf-8")
	if err != nil {
		return nil, nselib.NewDataNotFoundError("open turnover xls: " + err.Error())
	}
	sheet := wb.GetSheet(0)
	if sheet == nil {
		return nil, nselib.NewDataNotFoundError("turnover xls: no sheet")
	}
	var rows [][]string
	for r := 0; r <= int(sheet.MaxRow); r++ {
		row := sheet.Row(r)
		rec := make([]string, 4)
		for c := 0; c < 4; c++ {
			rec[c] = strings.TrimSpace(row.Col(c))
		}
		rows = append(rows, rec)
	}
	header := -1
	for i := 0; i < len(rows) && i < 16; i++ {
		if strings.ToLower(rows[i][0]) == "trade date" &&
			(strings.ToLower(rows[i][1]) == "category" || strings.ToLower(rows[i][1]) == "client categories") {
			header = i
			break
		}
	}
	if header < 0 {
		return nil, nselib.NewDataNotFoundError("category turnover FO data not found")
	}
	hdr := rows[header][:4]
	end := header + 1
	for end < len(rows) {
		f := strings.TrimSpace(rows[end][0])
		s := strings.ToLower(strings.TrimSpace(rows[end][1]))
		fl := strings.ToLower(f)
		if f == "" || strings.HasPrefix(fl, "note") || fl == "trade date" ||
			s == "category" || s == "client categories" {
			break
		}
		end++
	}
	rename := map[string]string{"Client Categories": "Category"}
	var out nselib.DataFrame
	for _, r := range rows[header+1 : end] {
		rec := nselib.Record{}
		for j, h := range hdr {
			if n, ok := rename[h]; ok {
				h = n
			}
			rec[h] = r[j]
		}
		cat, _ := rec["Category"].(string)
		if strings.TrimSpace(cat) == "" {
			continue
		}
		rec["Category"] = strings.TrimSpace(cat)
		buy := toFloat(rec["Buy Value in Rs.Crores"])
		sell := toFloat(rec["Sell Value in Rs.Crores"])
		rec["Buy Value in Rs.Crores"] = buy
		rec["Sell Value in Rs.Crores"] = sell
		rec["Net Value in Rs.Crores"] = buy - sell
		out = append(out, rec)
	}
	if out == nil {
		out = nselib.DataFrame{}
	}
	return out, nil
}

func toFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case string:
		f, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(t), ",", ""), 64)
		return f
	default:
		return 0
	}
}

// fiiStatsColumns mirrors the Python-assigned column names.
var fiiStatsColumns = []string{
	"fii_derivatives", "buy_contracts", "buy_value_in_Cr", "sell_contracts",
	"sell_value_in_Cr", "open_contracts", "open_contracts_value_in_Cr",
}

// FIIDerivativesStatistics fetches FII stats (.xls, skiprows=3, skipfooter=10).
func FIIDerivativesStatistics(tradeDate string) (nselib.DataFrame, error) {
	t, err := time.Parse(nselib.LayoutDDMMYYYY, tradeDate)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://nsearchives.nseindia.com/content/fo/fii_stats_%s.xls",
		t.Format(nselib.LayoutDDMMMYYYY))
	body, err := defaultClient.FetchBytes(url, "")
	if err != nil {
		return nil, nselib.NewDataNotFoundError(fmt.Sprintf("No data available for : %s", tradeDate))
	}
	wb, err := xls.OpenReader(bytes.NewReader(body), "utf-8")
	if err != nil {
		return nil, nselib.NewDataNotFoundError("open fii stats xls: " + err.Error())
	}
	sheet := wb.GetSheet(0)
	if sheet == nil {
		return nil, nselib.NewDataNotFoundError("fii stats xls: no sheet")
	}
	var rows [][]string
	for r := 0; r <= int(sheet.MaxRow); r++ {
		row := sheet.Row(r)
		rec := make([]string, len(fiiStatsColumns))
		for c := range rec {
			rec[c] = strings.TrimSpace(row.Col(c))
		}
		rows = append(rows, rec)
	}
	// skiprows=3, skipfooter=10, dropna
	if len(rows) > 3 {
		rows = rows[3:]
	}
	if len(rows) > 10 {
		rows = rows[:len(rows)-10]
	}
	var out nselib.DataFrame
	for _, r := range rows {
		empty := true
		for _, v := range r {
			if v != "" {
				empty = false
				break
			}
		}
		if empty {
			continue
		}
		rec := nselib.Record{}
		for i, c := range fiiStatsColumns {
			rec[c] = r[i]
		}
		out = append(out, rec)
	}
	if out == nil {
		out = nselib.DataFrame{}
	}
	return out, nil
}

// FNOSecurityInBanPeriod returns symbols banned from F&O for a trade date.
func FNOSecurityInBanPeriod(tradeDate string) ([]string, error) {
	t, err := time.Parse(nselib.LayoutDDMMYYYY, tradeDate)
	if err != nil {
		return nil, err
	}
	primary := fmt.Sprintf("https://nsearchives.nseindia.com/archives/fo/sec_ban/fo_secban_%s.csv",
		t.Format(nselib.LayoutDDMMYYYYCompact))
	body, err := defaultClient.FetchBytes(primary, "")
	if err != nil {
		fallback := "https://www.nseindia.com/api/reports?archives=" +
			"%5B%7B%22name%22%3A%22F%26O%20-%20Security%20in%20ban%20period%22%2C%22type%22%3A%22archives%22%2C%22category%22" +
			fmt.Sprintf("%%3A%%22derivatives%%22%%2C%%22section%%22%%3A%%22equity%%22%%7D%%5D&date=%s",
				t.Format(nselib.LayoutDDMMMYYYY)) + "&type=equity&mode=single"
		if body, err = defaultClient.FetchBytes(fallback, ""); err != nil {
			return nil, nselib.NewDataNotFoundError("Data not found, change the date...")
		}
	}
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	return parseBanLines(lines), nil
}

// parseBanLines extracts the symbol column (index 1) skipping the header.
func parseBanLines(lines []string) []string {
	var out []string
	for _, line := range lines[1:] {
		parts := strings.Split(line, ",")
		if len(parts) > 1 {
			out = append(out, strings.TrimSpace(parts[1]))
		}
	}
	return out
}

// LiveMostActiveUnderlying returns live most-active underlyings.
func LiveMostActiveUnderlying() (nselib.DataFrame, error) {
	var raw map[string]interface{}
	if err := defaultClient.FetchJSON("https://www.nseindia.com/api/live-analysis-most-active-underlying",
		"https://www.nseindia.com/market-data/most-active-underlying", &raw); err != nil {
		return nil, nselib.NewAPIError("Resource not available MSG: " + err.Error())
	}
	return recordsFromAny(raw["data"]), nil
}
