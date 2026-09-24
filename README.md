# nselib-go

A Go module for fetching publicly available data from [NSE India](https://www.nseindia.com).
Port of the Python [nselib](https://github.com/RuchiTanmay/nselib) library with full API parity,
strong typing (`DataFrame` = `[]Record`), shared cookie-primed HTTP client, and native concurrency.

## Installation

Requires Go 1.21+.

```bash
go get github.com/KartikeyaKotkar/nselib-go
```

## Quick start

```go
package main

import (
	"fmt"

	"github.com/KartikeyaKotkar/nselib-go/capitalmarket"
)

func main() {
	df, err := capitalmarket.PriceVolumeData("SBIN", "01-01-2024", "10-01-2024", "")
	if err != nil {
		panic(err)
	}
	fmt.Println(len(df), "rows")
}
```

Date ranges accept `from_date`/`to_date` in `dd-mm-YYYY` or a `period` (`1D`, `1W`, `1M`, `3M`, `6M`, `1Y`).
History functions chunk large ranges automatically (365 days for equities, 90 for F&O).

## Packages

| Package | Contents |
|---|---|
| `nselib` (root) | `DataFrame`/`Record` types, error hierarchy, HTTP client, CSV/date utilities, trading calendar |
| `capitalmarket` | Price/volume/deliverable history, bhavcopies, VaR, volatility, PE, bonds, live market, XBRL financials, turnover, CM growth |
| `derivatives` | Futures/options history, F&O bhavcopy, option chain (full/compact), participant OI/volume, FII stats, ban list, FO growth |
| `indices` | 107 indices across 4 categories, constituent lists, live performances |
| `debt` | WDM securities available for trading |
| `mutualfunds` | AMFI monthly archive scraping (xls/xlsx/html/pdf-text) |
| `nsdlfpi` | NSDL FPI investment/derivative activity (HTTP + headless-Chrome fallback) |
| `cashmarket` | Facade delegating to `mutualfunds` and `nsdlfpi` |

## Notes

- `nsdlfpi` browser automation needs a Chrome/Chromium executable; without one it uses the plain-HTTP fallback.
- AMFI PDF parsing is text-only (no `pdfplumber`-style table extraction); Excel/HTML reports keep full structure.
- Legacy `.xls` files parse via `github.com/extrame/xls`; `.xlsx` via `excelize`.

## Development

```bash
go build ./...
go vet ./...
go test ./... -race
```

## License

See [LICENSE](LICENSE).
