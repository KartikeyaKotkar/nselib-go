// Package capitalmarket ports nselib/capital_market to Go.
package capitalmarket

import (
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

// defaultClient is the shared NSE HTTP client. Override in tests via SetClient.
var defaultClient = nselib.NewNSEClient()

// SetClient overrides the HTTP client (tests).
func SetClient(c *nselib.NSEClient) { defaultClient = c }

// DateParams mirrors the Python from_date/to_date/period triple.
type DateParams struct {
	FromDate string
	ToDate   string
	Period   string
}

// resolveDateRange validates, derives and parses the date range.
func resolveDateRange(fromDate, toDate, period string) (time.Time, time.Time, error) {
	if err := nselib.ValidateDateParams(fromDate, toDate, period); err != nil {
		return time.Time{}, time.Time{}, err
	}
	fd, td, err := nselib.DeriveFromAndToDate(fromDate, toDate, period)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	from, err := time.Parse(nselib.LayoutDDMMYYYY, fd)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := time.Parse(nselib.LayoutDDMMYYYY, td)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return from, to, nil
}

// fetchInChunks breaks a date range into maxDays windows and concatenates results.
// Mirrors the Python 365-day while loop used by all history functions.
func fetchInChunks(maxDays int, from, to time.Time, fetcher func(fromStr, toStr string) (nselib.DataFrame, error)) (nselib.DataFrame, error) {
	var result nselib.DataFrame
	cur := from
	for !cur.After(to) {
		end := cur.AddDate(0, 0, maxDays-1)
		if end.After(to) {
			end = to
		}
		chunk, err := fetcher(cur.Format(nselib.LayoutDDMMYYYY), end.Format(nselib.LayoutDDMMYYYY))
		if err != nil {
			return nil, err
		}
		result = append(result, chunk...)
		cur = end.AddDate(0, 0, 1)
	}
	if result == nil {
		result = nselib.DataFrame{}
	}
	return result, nil
}

// PriceVolumeDeliverableData fetches OHLCV + deliverable data for a symbol.
func PriceVolumeDeliverableData(symbol, fromDate, toDate, period string) (nselib.DataFrame, error) {
	from, to, err := resolveDateRange(fromDate, toDate, period)
	if err != nil {
		return nil, err
	}
	symbol = nselib.CleanNSESymbol(symbol)
	return fetchInChunks(365, from, to, func(fs, ts string) (nselib.DataFrame, error) {
		return getPriceVolumeDeliverable(symbol, fs, ts)
	})
}

// PriceVolumeData fetches OHLCV price volume data for a symbol.
func PriceVolumeData(symbol, fromDate, toDate, period string) (nselib.DataFrame, error) {
	from, to, err := resolveDateRange(fromDate, toDate, period)
	if err != nil {
		return nil, err
	}
	symbol = nselib.CleanNSESymbol(symbol)
	return fetchInChunks(365, from, to, func(fs, ts string) (nselib.DataFrame, error) {
		return getPriceVolume(symbol, fs, ts)
	})
}

// DeliverableData fetches deliverable position data for a symbol.
func DeliverableData(symbol, fromDate, toDate, period string) (nselib.DataFrame, error) {
	from, to, err := resolveDateRange(fromDate, toDate, period)
	if err != nil {
		return nil, err
	}
	symbol = nselib.CleanNSESymbol(symbol)
	return fetchInChunks(365, from, to, func(fs, ts string) (nselib.DataFrame, error) {
		return getDeliverable(symbol, fs, ts)
	})
}

// IndiaVIXData fetches India VIX history.
func IndiaVIXData(fromDate, toDate, period string) (nselib.DataFrame, error) {
	from, to, err := resolveDateRange(fromDate, toDate, period)
	if err != nil {
		return nil, err
	}
	return fetchInChunks(365, from, to, func(fs, ts string) (nselib.DataFrame, error) {
		return getIndiaVIX(fs, ts)
	})
}

// IndexData fetches historical index data (e.g. "NIFTY 50").
func IndexData(index, fromDate, toDate, period string) (nselib.DataFrame, error) {
	from, to, err := resolveDateRange(fromDate, toDate, period)
	if err != nil {
		return nil, err
	}
	return fetchInChunks(365, from, to, func(fs, ts string) (nselib.DataFrame, error) {
		return getIndex(index, fs, ts)
	})
}

// BulkDealData fetches bulk deals.
func BulkDealData(fromDate, toDate, period string) (nselib.DataFrame, error) {
	from, to, err := resolveDateRange(fromDate, toDate, period)
	if err != nil {
		return nil, err
	}
	return fetchInChunks(365, from, to, func(fs, ts string) (nselib.DataFrame, error) {
		return getBulkDeals(fs, ts)
	})
}

// BlockDealsData fetches block deals.
func BlockDealsData(fromDate, toDate, period string) (nselib.DataFrame, error) {
	from, to, err := resolveDateRange(fromDate, toDate, period)
	if err != nil {
		return nil, err
	}
	return fetchInChunks(365, from, to, func(fs, ts string) (nselib.DataFrame, error) {
		return getBlockDeals(fs, ts)
	})
}

// ShortSellingData fetches short selling data.
func ShortSellingData(fromDate, toDate, period string) (nselib.DataFrame, error) {
	from, to, err := resolveDateRange(fromDate, toDate, period)
	if err != nil {
		return nil, err
	}
	return fetchInChunks(365, from, to, func(fs, ts string) (nselib.DataFrame, error) {
		return getShortSelling(fs, ts)
	})
}

// CorporateActionsForEquity fetches corporate actions (fnoOnly scopes to F&O securities).
func CorporateActionsForEquity(fromDate, toDate, period string, fnoOnly bool) (nselib.DataFrame, error) {
	from, to, err := resolveDateRange(fromDate, toDate, period)
	if err != nil {
		return nil, err
	}
	return getCorporateActions(
		from.Format(nselib.LayoutDDMMYYYY), to.Format(nselib.LayoutDDMMYYYY), fnoOnly)
}

// EventCalendarForEquity fetches the equity event calendar.
func EventCalendarForEquity(fromDate, toDate, period string, fnoOnly bool) (nselib.DataFrame, error) {
	from, to, err := resolveDateRange(fromDate, toDate, period)
	if err != nil {
		return nil, err
	}
	return getEventCalendar(
		from.Format(nselib.LayoutDDMMYYYY), to.Format(nselib.LayoutDDMMYYYY), fnoOnly)
}

// EquityList returns all tradable equities (subset of columns like Python).
func EquityList() (nselib.DataFrame, error) {
	df, err := defaultClient.FetchCSV(
		"https://archives.nseindia.com/content/equities/EQUITY_L.csv",
		"https://nsewebsite-staging.nseindia.com")
	if err != nil {
		return nil, err
	}
	return selectColumns(df, []string{"SYMBOL", "NAMEOFCOMPANY", "SERIES", "DATEOFLISTING", "FACEVALUE"}), nil
}

// FNOEquityList returns derivative equities with lot sizes.
func FNOEquityList() (nselib.DataFrame, error) {
	return getUnderlyingList("UnderlyingList")
}

// FNOIndexList returns derivative indices with lot sizes.
func FNOIndexList() (nselib.DataFrame, error) {
	return getUnderlyingList("IndexList")
}

// Nifty50EquityList returns NIFTY 50 constituents.
func Nifty50EquityList() (nselib.DataFrame, error) {
	df, err := defaultClient.FetchCSV(
		"https://nsearchives.nseindia.com/content/indices/ind_nifty50list.csv", "")
	if err != nil {
		return nil, err
	}
	return selectColumns(df, []string{"CompanyName", "Industry", "Symbol"}), nil
}

// NiftyNext50EquityList returns NIFTY Next 50 constituents.
func NiftyNext50EquityList() (nselib.DataFrame, error) {
	return fetchNiftyList("https://archives.nseindia.com/content/indices/ind_niftynext50list.csv")
}

// NiftyMidcap150EquityList returns NIFTY Midcap 150 constituents.
func NiftyMidcap150EquityList() (nselib.DataFrame, error) {
	return fetchNiftyList("https://archives.nseindia.com/content/indices/ind_niftymidcap150list.csv")
}

// NiftySmallcap250EquityList returns NIFTY Smallcap 250 constituents.
func NiftySmallcap250EquityList() (nselib.DataFrame, error) {
	return fetchNiftyList("https://archives.nseindia.com/content/indices/ind_niftysmallcap250list.csv")
}

func fetchNiftyList(url string) (nselib.DataFrame, error) {
	df, err := defaultClient.FetchCSV(url, "")
	if err != nil {
		return nil, err
	}
	return selectColumns(df, []string{"CompanyName", "Industry", "Symbol"}), nil
}

// selectColumns projects rows to wanted columns, matching case/space-insensitively.
func selectColumns(df nselib.DataFrame, want []string) nselib.DataFrame {
	out := make(nselib.DataFrame, 0, len(df))
	for _, r := range df {
		norm := map[string]interface{}{}
		for k, v := range r {
			nk := ""
			for _, c := range k {
				if c != ' ' {
					nk += string(c)
				}
			}
			norm[nk] = v
			// also index upper-cased for matching
			up := ""
			for _, c := range nk {
				if c >= 'a' && c <= 'z' {
					up += string(c - 32)
				} else {
					up += string(c)
				}
			}
			norm[up] = v
		}
		rec := nselib.Record{}
		for _, w := range want {
			if v, ok := norm[w]; ok {
				rec[w] = v
			} else {
				// fall back to original key of same name
				rec[w] = norm[w]
			}
		}
		out = append(out, rec)
	}
	return out
}
