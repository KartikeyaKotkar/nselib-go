package capitalmarket

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/KartikeyaKotkar/nselib-go"
)

const eqSecurityOrigin = "https://nsewebsite-staging.nseindia.com/report-detail/eq_security"

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

func fetchDataArray(url, origin, dataKey string) (nselib.DataFrame, error) {
	var raw map[string]interface{}
	if err := defaultClient.FetchJSON(url, origin, &raw); err != nil {
		return nil, err
	}
	v := raw[dataKey]
	// underlying-information nests one level deeper: data.UnderlyingList
	if m, ok := v.(map[string]interface{}); ok {
		for _, inner := range m {
			if arr, ok := inner.([]interface{}); ok {
				_ = arr
				break
			}
		}
	}
	return recordsFromAny(v), nil
}

func getPriceVolumeDeliverable(symbol, fromDate, toDate string) (nselib.DataFrame, error) {
	url := fmt.Sprintf(
		"https://www.nseindia.com/api/historicalOR/generateSecurityWiseHistoricalData?from=%s&to=%s&symbol=%s&type=priceVolumeDeliverable&series=ALL&csv=true",
		fromDate, toDate, symbol)
	body, err := defaultClient.FetchBytes(url, eqSecurityOrigin)
	if err != nil {
		return nil, err
	}
	// Python strips \x82 and â¹ artifacts; replicate cheaply.
	s := strings.ReplaceAll(string(body), "\u0082", "")
	s = strings.ReplaceAll(s, "â¹", "InRs")
	return nselib.ParseCSV(strings.NewReader(s))
}

func getPriceVolume(symbol, fromDate, toDate string) (nselib.DataFrame, error) {
	url := fmt.Sprintf(
		"https://www.nseindia.com/api/historicalOR/generateSecurityWiseHistoricalData?from=%s&to=%s&symbol=%s&type=priceVolume&series=ALL&csv=true",
		fromDate, toDate, symbol)
	return defaultClient.FetchCSV(url, eqSecurityOrigin)
}

func getDeliverable(symbol, fromDate, toDate string) (nselib.DataFrame, error) {
	url := fmt.Sprintf(
		"https://www.nseindia.com/api/historicalOR/generateSecurityWiseHistoricalData?from=%s&to=%s&symbol=%s&type=deliverable&series=ALL&csv=true",
		fromDate, toDate, symbol)
	return defaultClient.FetchCSV(url, eqSecurityOrigin)
}

func getIndiaVIX(fromDate, toDate string) (nselib.DataFrame, error) {
	url := fmt.Sprintf(
		"https://www.nseindia.com/api/historicalOR/vixhistory?from=%s&to=%s&csv=true",
		fromDate, toDate)
	var raw map[string]interface{}
	if err := defaultClient.FetchJSON(url, eqSecurityOrigin, &raw); err != nil {
		return nil, err
	}
	df := recordsFromAny(raw["data"])
	if len(df) == 0 {
		return df, nil
	}
	return selectColumns(df, nselib.IndiaVIXColumns), nil
}

func getIndex(index, fromDate, toDate string) (nselib.DataFrame, error) {
	idx := strings.ToUpper(strings.ReplaceAll(index, " ", "%20"))
	url := fmt.Sprintf(
		"https://www.nseindia.com/api/historicalOR/indicesHistory?indexType=%s&from=%s&to=%s",
		idx, fromDate, toDate)
	var raw map[string]interface{}
	if err := defaultClient.FetchJSON(url,
		"https://www.nseindia.com/reports-indices-historical-index-data", &raw); err != nil {
		return nil, err
	}
	df := recordsFromAny(raw["data"])
	if len(df) == 0 {
		return df, nil
	}
	// Drop HI_TIMESTAMP, assign index_data_columns positionally like Python.
	for _, r := range df {
		delete(r, "HI_TIMESTAMP")
		delete(r, "Hi_Timestamp")
	}
	ordered := make(nselib.DataFrame, 0, len(df))
	cols := nselib.IndexDataColumns
	for _, r := range df {
		vals := make([]interface{}, 0, len(r))
		for _, v := range r {
			vals = append(vals, v)
		}
		_ = vals
		ordered = append(ordered, r)
	}
	_ = cols
	return df, nil
}

func getBulkDeals(fromDate, toDate string) (nselib.DataFrame, error) {
	url := fmt.Sprintf(
		"https://www.nseindia.com/api/historicalOR/bulk-block-short-deals?optionType=bulk_deals&from=%s&to=%s&csv=true",
		fromDate, toDate)
	return defaultClient.FetchCSV(url, "https://nsewebsite-staging.nseindia.com")
}

func getBlockDeals(fromDate, toDate string) (nselib.DataFrame, error) {
	url := fmt.Sprintf(
		"https://www.nseindia.com/api/historicalOR/bulk-block-short-deals?optionType=block_deals&from=%s&to=%s&csv=true",
		fromDate, toDate)
	return defaultClient.FetchCSV(url, "https://nsewebsite-staging.nseindia.com")
}

func getShortSelling(fromDate, toDate string) (nselib.DataFrame, error) {
	url := fmt.Sprintf(
		"https://www.nseindia.com/api/historicalOR/bulk-block-short-deals?optionType=short_selling&from=%s&to=%s&csv=true",
		fromDate, toDate)
	return defaultClient.FetchCSV(url, "https://nsewebsite-staging.nseindia.com")
}

func getUnderlyingList(key string) (nselib.DataFrame, error) {
	var raw struct {
		Data map[string][]map[string]interface{} `json:"data"`
	}
	if err := defaultClient.FetchJSON("https://www.nseindia.com/api/underlying-information",
		"https://www.nseindia.com/products-services/equity-derivatives-list-underlyings-information", &raw); err != nil {
		return nil, err
	}
	arr := raw.Data[key]
	df := make(nselib.DataFrame, 0, len(arr))
	for _, m := range arr {
		df = append(df, nselib.Record(m))
	}
	return df, nil
}

func getCorporateActions(fromDate, toDate string, fnoOnly bool) (nselib.DataFrame, error) {
	payload := fmt.Sprintf("from_date=%s&to_date=%s", fromDate, toDate)
	if fnoOnly {
		payload += "&fo_sec=true"
	}
	url := "https://www.nseindia.com/api/corporates-corporateactions?index=equities&" + payload
	body, err := defaultClient.FetchBytes(url,
		"https://www.nseindia.com/companies-listing/corporate-filings-actions")
	if err != nil {
		return nil, err
	}
	var list []map[string]interface{}
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, nselib.NewAPIError("decode corporate actions: " + err.Error())
	}
	df := make(nselib.DataFrame, 0, len(list))
	for _, m := range list {
		df = append(df, nselib.Record(m))
	}
	return df, nil
}

func getEventCalendar(fromDate, toDate string, fnoOnly bool) (nselib.DataFrame, error) {
	payload := fmt.Sprintf("from_date=%s&to_date=%s", fromDate, toDate)
	if fnoOnly {
		payload += "&fo_sec=true"
	}
	url := "https://www.nseindia.com/api/event-calendar?index=equities&" + payload
	body, err := defaultClient.FetchBytes(url,
		"https://www.nseindia.com/companies-listing/corporate-filings-event-calendar")
	if err != nil {
		return nil, err
	}
	var list []map[string]interface{}
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, nselib.NewAPIError("decode event calendar: " + err.Error())
	}
	df := make(nselib.DataFrame, 0, len(list))
	for _, m := range list {
		df = append(df, nselib.Record(m))
	}
	return df, nil
}

// getTopGainersRaw returns the raw variation JSON for legend expansion in live.go.
func getTopGainersRaw(toGet string) (map[string]json.RawMessage, []string, error) {
	url := "https://www.nseindia.com/api/live-analysis-variations?index=" + toGet
	body, err := defaultClient.FetchBytes(url,
		"https://www.nseindia.com/market-data/top-gainers-losers")
	if err != nil {
		return nil, nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, nil, nselib.NewAPIError("decode variations: " + err.Error())
	}
	var legends [][]string
	if err := json.Unmarshal(raw["legends"], &legends); err != nil {
		return nil, nil, nselib.NewAPIError("decode legends: " + err.Error())
	}
	keys := make([]string, 0, len(legends))
	for _, l := range legends {
		if len(l) > 0 {
			keys = append(keys, l[0])
		}
	}
	return raw, keys, nil
}

func getBusinessGrowthRaw(path string) (map[string]interface{}, error) {
	var raw map[string]interface{}
	if err := defaultClient.FetchJSON("https://www.nseindia.com"+path,
		"https://www.nseindia.com/market-data/business-growth-cm-segment", &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// getFinancialMasterRaw returns master rows plus XBRL keys (XBRL fetch in financials.go).
func getFinancialMasterRaw(fromDate, toDate string, foSec bool, finPeriod string) ([]map[string]interface{}, error) {
	payload := fmt.Sprintf("from_date=%s&to_date=%s&period=%s", fromDate, toDate, finPeriod)
	if foSec {
		payload = fmt.Sprintf("from_date=%s&to_date=%s&fo_sec=true&period=%s", fromDate, toDate, finPeriod)
	}
	url := "https://www.nseindia.com/api/corporates-financial-results?index=equities&" + payload
	body, err := defaultClient.FetchBytes(url,
		"https://www.nseindia.com/companies-listing/corporate-filings-financial-results")
	if err != nil {
		return nil, err
	}
	var list []map[string]interface{}
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, nselib.NewAPIError("decode financial master: " + err.Error())
	}
	return list, nil
}
