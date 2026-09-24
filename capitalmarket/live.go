package capitalmarket

import (
	"encoding/json"
	"fmt"

	"github.com/KartikeyaKotkar/nselib-go"
)

// TotalTradedStocksResult holds the (dict, DataFrame) Python return as a struct.
type TotalTradedStocksResult struct {
	Summary map[string]interface{}
	Details nselib.DataFrame
}

// TopGainersOrLosers returns live gainers ("gainers") or losers ("loosers",sic Python spelling).
func TopGainersOrLosers(toGet string) (nselib.DataFrame, error) {
	if err := nselib.ValidateParamFromList(toGet, []string{"gainers", "loosers"}); err != nil {
		return nil, err
	}
	raw, keys, err := getTopGainersRaw(toGet)
	if err != nil {
		return nil, err
	}
	var out nselib.DataFrame
	for _, k := range keys {
		var wrap struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(raw[k], &wrap); err != nil {
			continue
		}
		for _, m := range wrap.Data {
			rec := nselib.Record(m)
			rec["legend"] = k
			out = append(out, rec)
		}
	}
	return out, nil
}

// MostActiveEquities fetches most active equities by volume/value.
func MostActiveEquities(fetchBy string) (nselib.DataFrame, error) {
	if err := nselib.ValidateParamFromList(fetchBy, []string{"volume", "value"}); err != nil {
		return nil, err
	}
	url := "https://www.nseindia.com/api/live-analysis-most-active-securities?index=" + fetchBy
	var raw map[string]interface{}
	if err := defaultClient.FetchJSON(url,
		"https://www.nseindia.com/market-data/most-active-equities", &raw); err != nil {
		return nil, err
	}
	return recordsFromAny(raw["data"]), nil
}

// TotalTradedStocks returns live market summary + details.
func TotalTradedStocks() (TotalTradedStocksResult, error) {
	var raw map[string]interface{}
	if err := defaultClient.FetchJSON("https://www.nseindia.com/api/live-analysis-stocksTraded",
		"https://www.nseindia.com/market-data/stocks-traded", &raw); err != nil {
		return TotalTradedStocksResult{}, err
	}
	total, _ := raw["total"].(map[string]interface{})
	if total == nil {
		return TotalTradedStocksResult{}, nselib.NewDataNotFoundError("total traded stocks: missing total")
	}
	summary, _ := total["count"].(map[string]interface{})
	return TotalTradedStocksResult{
		Summary: summary,
		Details: recordsFromAny(total["data"]),
	}, nil
}

// MarketWatchAllIndices returns all-indices snapshot.
func MarketWatchAllIndices() (nselib.DataFrame, error) {
	var raw map[string]interface{}
	if err := defaultClient.FetchJSON("https://www.nseindia.com/api/allIndices",
		"https://nsewebsite-staging.nseindia.com", &raw); err != nil {
		return nil, err
	}
	df := recordsFromAny(raw["data"])
	return selectColumns(df, []string{"key", "index", "indexSymbol", "last", "variation",
		"percentChange", "open", "high", "low", "previousClose", "yearHigh", "yearLow",
		"pe", "pb", "dy", "declines", "advances", "unchanged", "perChange365d",
		"perChange30d", "previousDay", "oneWeekAgoVal", "oneMonthAgoVal", "oneYearAgoVal"}), nil
}

// FIIDIITradingActivity returns FII/DII activity of the day.
func FIIDIITradingActivity() (nselib.DataFrame, error) {
	body, err := defaultClient.FetchBytes("https://www.nseindia.com/api/fiidiiTradeReact", "")
	if err != nil {
		return nil, err
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal(body, &arr); err != nil {
		// some responses wrap in object; fall back to generic
		var raw interface{}
		if err2 := json.Unmarshal(body, &raw); err2 != nil {
			return nil, nselib.NewAPIError(fmt.Sprintf("decode fii/dii: %v", err))
		}
		if m, ok := raw.(map[string]interface{}); ok {
			return recordsFromAny(m["data"]), nil
		}
		return nil, nselib.NewAPIError("decode fii/dii: " + err.Error())
	}
	df := make(nselib.DataFrame, 0, len(arr))
	for _, m := range arr {
		df = append(df, nselib.Record(m))
	}
	return df, nil
}
