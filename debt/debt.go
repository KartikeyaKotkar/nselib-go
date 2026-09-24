// Package debt ports nselib/debt to Go.
package debt

import (
	"fmt"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

var defaultClient = nselib.NewNSEClient()

// SetClient overrides the HTTP client (tests).
func SetClient(c *nselib.NSEClient) { defaultClient = c }

// SecuritiesAvailableForTrading fetches WDM securities for a trade date (dd-mm-YYYY).
func SecuritiesAvailableForTrading(tradeDate string) (nselib.DataFrame, error) {
	t, err := time.Parse(nselib.LayoutDDMMYYYY, tradeDate)
	if err != nil {
		return nil, err
	}
	month := t.Format("Jan")
	upper := ""
	for _, c := range month {
		if c >= 'a' && c <= 'z' {
			upper += string(c - 32)
		} else {
			upper += string(c)
		}
	}
	url := fmt.Sprintf("https://nsearchives.nseindia.com/content/historical/WDM/%s/%s/wdmlist_%s.csv",
		t.Format("2006"), upper, t.Format(nselib.LayoutDDMMYYYYCompact))
	return defaultClient.FetchCSV(url, "https://www.nseindia.com/all-reports-debt")
}
