package capitalmarket

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
	"github.com/extrame/xls"
)

// CategoryTurnoverCash fetches cash-market category turnover (.xls BIFF8 via extrame/xls).
func CategoryTurnoverCash(tradeDate string) (nselib.DataFrame, error) {
	t, err := time.Parse(nselib.LayoutDDMMYYYY, tradeDate)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://archives.nseindia.com/archives/equities/cat/cat_turnover_%s.xls",
		t.Format(nselib.LayoutDDMMYYCompact))
	body, err := defaultClient.FetchBytes(url, "")
	if err != nil {
		return nil, err
	}
	return parseCategoryTurnoverXLS(body)
}

// parseCategoryTurnoverXLS replicates the Python block-scan over xlrd rows.
func parseCategoryTurnoverXLS(body []byte) (nselib.DataFrame, error) {
	wb, err := xls.OpenReader(bytes.NewReader(body), "utf-8")
	if err != nil {
		return nil, nselib.NewDataNotFoundError("open turnover xls: " + err.Error())
	}
	sheet := wb.GetSheet(0)
	if sheet == nil {
		return nil, nselib.NewDataNotFoundError("turnover xls: no sheet")
	}
	// materialize first 4 columns as strings
	var rows [][]string
	for r := 0; r <= int(sheet.MaxRow); r++ {
		row := sheet.Row(r)
		rec := make([]string, 4)
		for c := 0; c < 4; c++ {
			if col := row.Col(c); col != "" {
				rec[c] = strings.TrimSpace(col)
			}
		}
		rows = append(rows, rec)
	}
	type block struct {
		header []string
		rows   [][]string
	}
	var blocks []block
	i := 0
	for i < len(rows) {
		c0 := strings.ToLower(rows[i][0])
		c1 := strings.ToLower(rows[i][1])
		if c0 == "trade date" && (c1 == "category" || c1 == "client categories") {
			hdr := append([]string{}, rows[i][:4]...)
			start := i + 1
			end := start
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
			blocks = append(blocks, block{header: hdr, rows: rows[start:end]})
			i = end
			continue
		}
		i++
	}
	if len(blocks) == 0 {
		return nil, nselib.NewDataNotFoundError("category turnover cash data not found")
	}
	rename := map[string]string{
		"Client Categories": "Category", "Buy Value in Rs.": "Buy Value in Rs.Crores",
		"Sell Value in Rs.": "Sell Value in Rs.Crores",
	}
	var out nselib.DataFrame
	for _, b := range blocks {
		hdr := make([]string, len(b.header))
		for j, h := range b.header {
			if n, ok := rename[h]; ok {
				hdr[j] = n
			} else {
				hdr[j] = h
			}
		}
		for _, r := range b.rows {
			rec := nselib.Record{}
			for j, h := range hdr {
				if h == "" {
					continue
				}
				rec[h] = r[j]
			}
			out = append(out, rec)
		}
	}
	// normalize: keep 4 cols, drop empty category, coerce numerics, add net
	final := make(nselib.DataFrame, 0, len(out))
	for _, r := range out {
		cat, _ := r["Category"].(string)
		cat = strings.TrimSpace(cat)
		if cat == "" {
			continue
		}
		td, _ := r["Trade Date"].(string)
		if td == "" {
			continue
		}
		buy := toFloat(r["Buy Value in Rs.Crores"])
		sell := toFloat(r["Sell Value in Rs.Crores"])
		final = append(final, nselib.Record{
			"Trade Date":              td,
			"Category":                cat,
			"Buy Value in Rs.Crores":  buy,
			"Sell Value in Rs.Crores": sell,
			"Net Value in Rs.Crores":  buy - sell,
		})
	}
	return final, nil
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
