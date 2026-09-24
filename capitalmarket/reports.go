package capitalmarket

import (
	"bytes"
	"fmt"

	"github.com/KartikeyaKotkar/nselib-go"
)

// varReport fetches a C_VAR1 .DAT file and assigns VarColumns (skiprows=1 like Python).
func varReport(tradeDate string, slot int) (nselib.DataFrame, error) {
	t, err := parseTradeDate(tradeDate)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://nsearchives.nseindia.com/archives/nsccl/var/C_VAR1_%s_%d.DAT",
		t.Format(nselib.LayoutDDMMYYYYCompact), slot)
	body, err := defaultClient.FetchBytes(url, "")
	if err != nil {
		return nil, err
	}
	df, err := nselib.ParseCSV(bytes.NewReader(body), nselib.SkipRows(1))
	if err != nil {
		return nil, err
	}
	// Reassign positional VarColumns like Python (columns = var_columns).
	out := make(nselib.DataFrame, 0, len(df))
	for _, r := range df {
		vals := make([]interface{}, 0, len(r))
		for _, v := range r {
			vals = append(vals, v)
		}
		rec := nselib.Record{}
		i := 0
		for _, c := range nselib.VarColumns {
			if i < len(vals) {
				rec[c] = vals[i]
			} else {
				rec[c] = ""
			}
			i++
		}
		out = append(out, rec)
	}
	return out, nil
}

// VarBeginDay fetches VaR begin-of-day.
func VarBeginDay(tradeDate string) (nselib.DataFrame, error) { return varReport(tradeDate, 1) }

// Var1stIntraDay fetches VaR 1st intra-day.
func Var1stIntraDay(tradeDate string) (nselib.DataFrame, error) { return varReport(tradeDate, 2) }

// Var2ndIntraDay fetches VaR 2nd intra-day.
func Var2ndIntraDay(tradeDate string) (nselib.DataFrame, error) { return varReport(tradeDate, 3) }

// Var3rdIntraDay fetches VaR 3rd intra-day.
func Var3rdIntraDay(tradeDate string) (nselib.DataFrame, error) { return varReport(tradeDate, 4) }

// Var4thIntraDay fetches VaR 4th intra-day.
func Var4thIntraDay(tradeDate string) (nselib.DataFrame, error) { return varReport(tradeDate, 5) }

// VarEndOfDay fetches VaR end-of-day.
func VarEndOfDay(tradeDate string) (nselib.DataFrame, error) { return varReport(tradeDate, 6) }

// DailyVolatility fetches CM daily volatility with 3-URL fallback like Python.
func DailyVolatility(tradeDate string) (nselib.DataFrame, error) {
	t, err := parseTradeDate(tradeDate)
	if err != nil {
		return nil, err
	}
	urls := []string{
		fmt.Sprintf("https://nsearchives.nseindia.com/archives/nsccl/volt/CMVOLT_%s.CSV",
			t.Format(nselib.LayoutDDMMYYYYCompact)),
		fmt.Sprintf("https://archives.nseindia.com/archives/nsccl/volt/CMVOLT_%s.CSV",
			t.Format(nselib.LayoutDDMMYYYYCompact)),
		"https://www.nseindia.com/api/reports?archives=" +
			"%5B%7B%22name%22%3A%22CM%20-%20Daily%20Volatility%22%2C%22type%22%3A%22archives%22%2C%22category%22" +
			fmt.Sprintf("%%3A%%22capital-market%%22%%2C%%22section%%22%%3A%%22equity%%22%%7D%%5D&date=%s",
				t.Format(nselib.LayoutDDMMMYYYY)) + "&type=equity&mode=single",
	}
	var lastErr error
	for _, u := range urls {
		body, err := defaultClient.FetchBytes(u, "https://www.nseindia.com/all-reports")
		if err != nil {
			lastErr = err
			continue
		}
		df, err := nselib.ParseCSV(bytes.NewReader(body))
		if err != nil {
			lastErr = err
			continue
		}
		// drop all-empty rows like Python dropna(how='all')
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
		lastErr = nselib.NewDataNotFoundError("no CM daily volatility data for " + tradeDate)
	}
	return nil, lastErr
}

// SMEBandComplete fetches SME price band file.
func SMEBandComplete(tradeDate string) (nselib.DataFrame, error) {
	t, err := parseTradeDate(tradeDate)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://nsearchives.nseindia.com/sme/content/price_band/archieves/sme_bands_complete_%s.csv",
		t.Format(nselib.LayoutDDMMYYYYCompact))
	return defaultClient.FetchCSV(url, "")
}

// Week52HighLowReport fetches 52-week high/low (skiprows=2 like Python).
func Week52HighLowReport(tradeDate string) (nselib.DataFrame, error) {
	t, err := parseTradeDate(tradeDate)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://nsearchives.nseindia.com/content/CM_52_wk_High_low_%s.csv",
		t.Format(nselib.LayoutDDMMYYYYCompact))
	body, err := defaultClient.FetchBytes(url, "")
	if err != nil {
		return nil, err
	}
	return nselib.ParseCSV(bytes.NewReader(body), nselib.SkipRows(2))
}

// CorporateBondTradeReport fetches corporate bond trades.
func CorporateBondTradeReport(tradeDate string) (nselib.DataFrame, error) {
	t, err := parseTradeDate(tradeDate)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://nsearchives.nseindia.com/archives/equities/corpbond/corpbond%s.csv",
		t.Format(nselib.LayoutDDMMYYCompact))
	return defaultClient.FetchCSV(url, "")
}

// PERatio fetches PE ratios for all equities.
func PERatio(tradeDate string) (nselib.DataFrame, error) {
	t, err := parseTradeDate(tradeDate)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://nsearchives.nseindia.com/content/equities/peDetail/PE_%s.csv",
		t.Format(nselib.LayoutDDMMYYCompact))
	return defaultClient.FetchCSV(url, "")
}
