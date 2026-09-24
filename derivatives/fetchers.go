package derivatives

import (
	"fmt"
	"strings"

	"github.com/KartikeyaKotkar/nselib-go"
)

const foSecurityOrigin = "https://www.nseindia.com/report-detail/fo_eq_security"

// recordsFromAny converts JSON arrays of objects into a DataFrame.
func recordsFromAny(v interface{}) nselib.DataFrame {
	arr, ok := v.([]interface{})
	if !ok {
		return nselib.DataFrame{}
	}
	df := make(nselib.DataFrame, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		df = append(df, nselib.Record(m))
	}
	return df
}

func getFuture(symbol, instrument, fromDate, toDate string) (nselib.DataFrame, error) {
	url := fmt.Sprintf(
		"https://www.nseindia.com/api/historicalOR/foCPV?from=%s&to=%s&instrumentType=%s&symbol=%s&csv=true",
		fromDate, toDate, instrument, symbol)
	var raw map[string]interface{}
	if err := defaultClient.FetchJSON(url, foSecurityOrigin, &raw); err != nil {
		return nil, nselib.NewDataNotFoundError("Invalid parameters : NSE error:" + err.Error())
	}
	df := recordsFromAny(raw["data"])
	for _, r := range df {
		for k, v := range r {
			nk := strings.ReplaceAll(k, "FH_", "")
			nk = strings.ReplaceAll(nk, "EOD_", "")
			nk = strings.ReplaceAll(nk, "HIT_", "")
			if nk != k {
				r[nk] = v
				delete(r, k)
			}
		}
	}
	return df, nil
}

func getOption(symbol, instrument, optionType, fromDate, toDate string) (nselib.DataFrame, error) {
	url := fmt.Sprintf(
		"https://www.nseindia.com/api/historicalOR/foCPV?from=%s&to=%s&instrumentType=%s&symbol=%s&optionType=%s&csv=true",
		fromDate, toDate, instrument, symbol, optionType)
	var raw map[string]interface{}
	if err := defaultClient.FetchJSON(url, foSecurityOrigin, &raw); err != nil {
		return nil, nselib.NewDataNotFoundError("Invalid parameters : NSE error :" + err.Error())
	}
	df := recordsFromAny(raw["data"])
	if len(df) == 0 {
		return nil, nselib.NewDataNotFoundError("Invalid parameters, Please change the parameters")
	}
	return selectColumns(df, nselib.FuturePriceVolumeColumns), nil
}

// selectColumns projects rows to wanted columns present in the row.
func selectColumns(df nselib.DataFrame, want []string) nselib.DataFrame {
	out := make(nselib.DataFrame, 0, len(df))
	for _, r := range df {
		rec := nselib.Record{}
		for _, w := range want {
			if v, ok := r[w]; ok {
				rec[w] = v
			}
		}
		out = append(out, rec)
	}
	return out
}

// isIndexSymbol reports whether a symbol is an index underlying.
func isIndexSymbol(symbol string) bool {
	for _, ind := range nselib.IndicesList {
		if strings.Contains(symbol, ind) {
			return true
		}
	}
	return false
}

// getOptionChainRaw fetches the raw option-chain JSON. Empty expiry means all expiries.
func getOptionChainRaw(symbol, expiryDDMonYYYY string) (map[string]interface{}, error) {
	symbol = nselib.CleanNSESymbol(symbol)
	typ := "Equity"
	if isIndexSymbol(symbol) {
		typ = "Indices"
	}
	url := fmt.Sprintf("https://www.nseindia.com/api/option-chain-v3?type=%s&symbol=%s&expiry=%s",
		typ, symbol, expiryDDMonYYYY)
	var raw map[string]interface{}
	if err := defaultClient.FetchJSON(url, "https://www.nseindia.com/option-chain", &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func getBusinessGrowthRaw(path string) (map[string]interface{}, error) {
	var raw map[string]interface{}
	if err := defaultClient.FetchJSON("https://www.nseindia.com"+path,
		"https://www.nseindia.com/market-data/business-growth-fo-segment", &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
