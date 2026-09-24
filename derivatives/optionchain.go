package derivatives

import (
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

// CompactColumns is the oi_mode=compact column order from Python.
var CompactColumns = []string{
	"Fetch_Time", "Symbol", "Expiry_Date",
	"CALLS_OI", "CALLS_Chng_in_OI", "CALLS_Volume", "CALLS_IV", "CALLS_LTP", "CALLS_Net_Chng",
	"Strike_Price",
	"PUTS_OI", "PUTS_Chng_in_OI", "PUTS_Volume", "PUTS_IV", "PUTS_LTP", "PUTS_Net_Chng",
}

// FullColumns is the oi_mode=full column order from Python.
var FullColumns = []string{
	"Fetch_Time", "Symbol", "Expiry_Date",
	"CALLS_OI", "CALLS_Chng_in_OI", "CALLS_Volume", "CALLS_IV", "CALLS_LTP", "CALLS_Net_Chng",
	"CALLS_Bid_Qty", "CALLS_Bid_Price", "CALLS_Ask_Price", "CALLS_Ask_Qty",
	"Strike_Price",
	"PUTS_Bid_Qty", "PUTS_Bid_Price", "PUTS_Ask_Price", "PUTS_Ask_Qty",
	"PUTS_Net_Chng", "PUTS_LTP", "PUTS_IV", "PUTS_Volume", "PUTS_Chng_in_OI", "PUTS_OI",
}

func num(m map[string]interface{}, key string) interface{} {
	if m == nil {
		return 0
	}
	if v, ok := m[key]; ok && v != nil {
		return v
	}
	return 0
}

// parseOptionChain converts raw chain JSON into rows filtered by expiry ("" = all).
func parseOptionChain(raw map[string]interface{}, symbol, expiryDDMonYYYY, oiMode string) nselib.DataFrame {
	records, _ := raw["records"].(map[string]interface{})
	timestamp, _ := records["timestamp"].(string)
	items, _ := records["data"].([]interface{})
	cols := FullColumns
	if oiMode == "compact" {
		cols = CompactColumns
	}
	var out nselib.DataFrame
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		exp, _ := m["expiryDate"].(string)
		if expiryDDMonYYYY != "" && exp != expiryDDMonYYYY {
			continue
		}
		ce, _ := m["CE"].(map[string]interface{})
		pe, _ := m["PE"].(map[string]interface{})
		strike := m["strikePrice"]
		row := nselib.Record{
			"Fetch_Time": timestamp, "Symbol": symbol, "Expiry_Date": exp,
			"CALLS_OI": num(ce, "openInterest"), "CALLS_Chng_in_OI": num(ce, "changeinOpenInterest"),
			"CALLS_Volume": num(ce, "totalTradedVolume"), "CALLS_IV": num(ce, "impliedVolatility"),
			"CALLS_LTP": num(ce, "lastPrice"), "CALLS_Net_Chng": num(ce, "change"),
			"CALLS_Bid_Qty": num(ce, "buyQuantity1"), "CALLS_Bid_Price": num(ce, "buyPrice1"),
			"CALLS_Ask_Price": num(ce, "sellPrice1"), "CALLS_Ask_Qty": num(ce, "sellQuantity1"),
			"Strike_Price": strike,
			"PUTS_OI":      num(pe, "openInterest"), "PUTS_Chng_in_OI": num(pe, "changeinOpenInterest"),
			"PUTS_Volume": num(pe, "totalTradedVolume"), "PUTS_IV": num(pe, "impliedVolatility"),
			"PUTS_LTP": num(pe, "lastPrice"), "PUTS_Net_Chng": num(pe, "change"),
			"PUTS_Bid_Qty": num(pe, "buyQuantity1"), "PUTS_Bid_Price": num(pe, "buyPrice1"),
			"PUTS_Ask_Price": num(pe, "sellPrice1"), "PUTS_Ask_Qty": num(pe, "sellQuantity1"),
		}
		ordered := nselib.Record{}
		for _, c := range cols {
			ordered[c] = row[c]
		}
		out = append(out, ordered)
	}
	if out == nil {
		out = nselib.DataFrame{}
	}
	return out
}

// NSELiveOptionChain fetches the live option chain.
// expiryDate is dd-mm-YYYY ("" = all expiries); oiMode is "full" or "compact".
func NSELiveOptionChain(symbol, expiryDate, oiMode string) (nselib.DataFrame, error) {
	if oiMode == "" {
		oiMode = "full"
	}
	if err := nselib.ValidateParamFromList(oiMode, []string{"full", "compact"}); err != nil {
		return nil, err
	}
	exp := ""
	if expiryDate != "" {
		t, err := time.Parse(nselib.LayoutDDMMYYYY, expiryDate)
		if err != nil {
			return nil, err
		}
		exp = t.Format(nselib.LayoutDDMMMYYYY)
	}
	raw, err := getOptionChainRaw(symbol, exp)
	if err != nil {
		return nil, err
	}
	return parseOptionChain(raw, symbol, exp, oiMode), nil
}
