// Package derivatives ports nselib/derivatives to Go.
package derivatives

import (
	"strings"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

// defaultClient is the shared NSE HTTP client. Override in tests via SetClient.
var defaultClient = nselib.NewNSEClient()

// SetClient overrides the HTTP client (tests).
func SetClient(c *nselib.NSEClient) { defaultClient = c }

// fetchInChunks breaks a date range into maxDays windows and concatenates results.
// F&O history uses 90-day windows (vs 365 for capital market).
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

// FuturePriceVolumeData fetches futures price & volume (instrument FUTIDX/FUTSTK).
func FuturePriceVolumeData(symbol, instrument, fromDate, toDate, period string) (nselib.DataFrame, error) {
	from, to, err := resolveDateRange(fromDate, toDate, period)
	if err != nil {
		return nil, err
	}
	symbol = nselib.CleanNSESymbol(symbol)
	instrument = strings.ToUpper(instrument)
	if instrument != "FUTIDX" && instrument != "FUTSTK" {
		return nil, nselib.NewDerivativeInstrumentNotFoundError(instrument + " is not a future instrument")
	}
	return fetchInChunks(90, from, to, func(fs, ts string) (nselib.DataFrame, error) {
		return getFuture(symbol, instrument, fs, ts)
	})
}

// OptionPriceVolumeData fetches options price & volume.
// Empty optionType fetches both PE and CE like Python.
func OptionPriceVolumeData(symbol, instrument, optionType, fromDate, toDate, period string) (nselib.DataFrame, error) {
	from, to, err := resolveDateRange(fromDate, toDate, period)
	if err != nil {
		return nil, err
	}
	symbol = nselib.CleanNSESymbol(symbol)
	instrument = strings.ToUpper(instrument)
	if instrument != "OPTIDX" && instrument != "OPTSTK" {
		return nil, nselib.NewDerivativeInstrumentNotFoundError(instrument + " is not an option instrument")
	}
	var types []string
	if optionType != "" {
		optionType = strings.ToUpper(optionType)
		if optionType != "PE" && optionType != "CE" {
			return nil, nselib.NewDerivativeInstrumentNotFoundError(optionType + " is not a valid option type")
		}
		types = []string{optionType}
	} else {
		types = []string{"PE", "CE"}
	}
	return fetchInChunks(90, from, to, func(fs, ts string) (nselib.DataFrame, error) {
		var out nselib.DataFrame
		for _, ot := range types {
			chunk, err := getOption(symbol, instrument, ot, fs, ts)
			if err != nil {
				return nil, err
			}
			out = append(out, chunk...)
		}
		return out, nil
	})
}

// ExpiryDatesFuture returns valid futures expiry dates.
func ExpiryDatesFuture() ([]string, error) {
	var payload struct {
		ExpiryDates []string `json:"expiryDates"`
	}
	if err := defaultClient.FetchJSON(
		"https://www.nseindia.com/api/option-chain-contract-info?symbol=TCS",
		"https://www.nseindia.com/option-chain", &payload); err != nil {
		return nil, err
	}
	return payload.ExpiryDates, nil
}

// ExpiryDatesOptionIndex maps each underlying index to its expiry dates.
func ExpiryDatesOptionIndex() (map[string][]string, error) {
	out := map[string][]string{}
	for _, ind := range nselib.IndicesList {
		var payload struct {
			ExpiryDates []string `json:"expiryDates"`
		}
		if err := defaultClient.FetchJSON(
			"https://www.nseindia.com/api/option-chain-contract-info?symbol="+ind,
			"https://www.nseindia.com/option-chain", &payload); err != nil {
			return nil, err
		}
		out[ind] = payload.ExpiryDates
	}
	return out, nil
}
