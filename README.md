<p align="center">
  <h1 align="center">nselib-go</h1>
  <p align="center">
    A Go module to fetch publicly available data from <a href="https://www.nseindia.com">NSE India</a>.
  </p>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/KartikeyaKotkar/nselib-go"><img src="https://pkg.go.dev/badge/github.com/KartikeyaKotkar/nselib-go.svg" alt="Go Reference"></a>
  <a href="https://github.com/KartikeyaKotkar/nselib-go/actions/workflows/ci.yml"><img src="https://github.com/KartikeyaKotkar/nselib-go/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/KartikeyaKotkar/nselib-go/releases"><img src="https://img.shields.io/github/v/release/KartikeyaKotkar/nselib-go?color=blue" alt="Release"></a>
  <a href="https://github.com/KartikeyaKotkar/nselib-go/blob/main/LICENSE"><img src="https://img.shields.io/github/license/KartikeyaKotkar/nselib-go" alt="License"></a>
</p>

<p align="center">
  Full-API port of the Python <a href="https://github.com/RuchiTanmay/nselib">nselib</a> library -
  strong typing (<code>DataFrame</code> = <code>[]Record</code>), a shared cookie-primed HTTP client,
  and concurrent fetching throughout.
</p>

---

## Features

- **Capital Market** - Price volume data, deliverable positions, bhav copies, bulk/block deals, short selling, VaR margins, PE ratios, 52-week highs/lows, and more
- **Cash Market** - NSDL FPI investment and derivative activity plus AMFI monthly archive reports
- **Derivatives** - Futures & options price volume data, bhav copies, participant-wise OI & volume, live option chains, FII statistics, ban period securities
- **Indices** - Index constituent lists, live index performances across Broad Market, Sectoral, Thematic, and Strategy categories
- **Debt** - Securities available for trading
- **Corporate Filings** - Financial results, corporate actions, event calendars
- **Market Activity** - Top gainers/losers, most active equities, total traded stocks, FII/DII activity
- **Utilities** - Trading holiday calendar, India VIX historical data

## Installation

Requires Go 1.21+ (recent toolchain recommended; dependencies resolve it automatically).

```bash
go get github.com/KartikeyaKotkar/nselib-go
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/KartikeyaKotkar/nselib-go/capitalmarket"
)

func main() {
	// Price volume data for a stock (last 1 month)
	df, err := capitalmarket.PriceVolumeData("SBIN", "", "", "1M")
	if err != nil {
		panic(err)
	}
	fmt.Println(len(df), "rows")

	// Or specify a custom date range
	df, err = capitalmarket.PriceVolumeAndDeliverablePositionData(
		"SBIN", "01-01-2024", "31-01-2024", "",
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(df[0])
}
```

Rows are `map[string]interface{}` (`nselib.Record`); a result set is `nselib.DataFrame` (`[]Record`).

## API Reference

### Date Parameters

Most functions accept dates in two ways:

| Parameter | Format | Example |
|---|---|---|
| `fromDate` / `toDate` | `dd-mm-YYYY` | `"01-06-2024"` |
| `period` | Shorthand code | `"1D"`, `"1W"`, `"1M"`, `"3M"`, `"6M"`, `"1Y"` |

> You must provide **either** `fromDate` + `toDate` **or** `period`, not both. History windows fetch concurrently and assemble in chronological order (365-day windows for equities, 90-day for F&O).

---

### Capital Market

```go
import "github.com/KartikeyaKotkar/nselib-go/capitalmarket"
```

| Function | Description | Key Parameters |
|---|---|---|
| `PriceVolumeAndDeliverablePositionData()` | OHLCV + delivery data | `symbol`, dates |
| `PriceVolumeData()` | OHLCV price volume data | `symbol`, dates |
| `DeliverablePositionData()` | Delivery position data | `symbol`, dates |
| `BulkDealData()` | Bulk deal transactions | dates |
| `BlockDealsData()` | Block deal transactions | dates |
| `ShortSellingData()` | Short selling reports | dates |
| `BhavCopyWithDelivery()` | Daily bhav copy with delivery | `tradeDate` |
| `BhavCopyEquities()` | CM-UDiFF bhav copy | `tradeDate` |
| `BhavCopyIndices()` | Index closing bhav copy | `tradeDate` |
| `BhavCopySME()` / `SMEBhavCopy()` | SME bhav copy | `tradeDate` |
| `EquityList()` | All listed equities | - |
| `FNOEquityList()` | F&O equity list with lot sizes | - |
| `FNOIndexList()` | F&O index list with lot sizes | - |
| `Nifty50EquityList()` | Nifty 50 constituents | - |
| `NiftyNext50EquityList()` | Nifty Next 50 constituents | - |
| `NiftyMidcap150EquityList()` | Nifty Midcap 150 constituents | - |
| `NiftySmallcap250EquityList()` | Nifty Smallcap 250 constituents | - |
| `IndiaVIXData()` | India VIX historical data | dates |
| `IndexData()` | Historical index OHLC data | `index`, dates |
| `MarketWatchAllIndices()` | Live snapshot of all indices | - |
| `DailyVolatility()` | CM daily volatility report | `tradeDate` |
| `FIIDIITradingActivity()` | FII/DII buy-sell activity | - |
| `VarBeginDay()` | VaR - begin of day | `tradeDate` |
| `Var1stIntraDay()` | VaR - 1st intraday | `tradeDate` |
| `Var2ndIntraDay()` | VaR - 2nd intraday | `tradeDate` |
| `Var3rdIntraDay()` | VaR - 3rd intraday | `tradeDate` |
| `Var4thIntraDay()` | VaR - 4th intraday | `tradeDate` |
| `VarEndOfDay()` | VaR - end of day | `tradeDate` |
| `SMEBandComplete()` | SME band complete data | `tradeDate` |
| `Week52HighLowReport()` | 52-week high/low report | `tradeDate` |
| `FinancialResultsForEquity()` | Quarterly/annual financials | dates, `foSec`, `finPeriod` |
| `CorporateBondTradeReport()` | Corporate bond trades | `tradeDate` |
| `PERatio()` | PE ratio for all equities | `tradeDate` |
| `CorporateActionsForEquity()` | Corporate actions | dates, `fnoOnly` |
| `EventCalendarForEquity()` | Event calendar | dates, `fnoOnly` |
| `TopGainersOrLosers()` | Top gainers or losers | `toGet` (`"gainers"` / `"loosers"`) |
| `MostActiveEquities()` | Most active by value/volume | `fetchBy` (`"value"` / `"volume"`) |
| `TotalTradedStocks()` | Traded stocks summary + details | - (returns `TotalTradedStocksResult`) |
| `CategoryTurnoverCash()` | Category-wise turnover data | `tradeDate` |
| `BusinessGrowthCMSegment()` | Business growth, CM segment | `dataType`, `fromYear`, `toYear`, `month`, `year` |

**Examples:**

```go
// Bhav copy for a specific date
df, _ := capitalmarket.BhavCopyWithDelivery("20-06-2024")

// India VIX for last 1 week
df, _ = capitalmarket.IndiaVIXData("", "", "1W")

// Historical index data
df, _ = capitalmarket.IndexData("NIFTY 50", "01-01-2024", "31-03-2024", "")

// Financial results (quarterly, F&O securities only)
df, _ = capitalmarket.FinancialResultsForEquity("", "", "6M", true, "Quarterly")

// Top gainers in live market
df, _ = capitalmarket.TopGainersOrLosers("gainers")
```

---

### Derivatives

```go
import "github.com/KartikeyaKotkar/nselib-go/derivatives"
```

| Function | Description | Key Parameters |
|---|---|---|
| `FuturePriceVolumeData()` | Futures price & volume | `symbol`, `instrument` (`FUTIDX`/`FUTSTK`), dates |
| `OptionPriceVolumeData()` | Options price & volume | `symbol`, `instrument` (`OPTIDX`/`OPTSTK`), `optionType` (`PE`/`CE`/"" for both), dates |
| `FNOBhavCopy()` | F&O daily bhav copy | `tradeDate` |
| `ParticipantWiseOpenInterest()` | OI by participant category | `tradeDate` |
| `ParticipantWiseTradingVolume()` | Volume by participant category | `tradeDate` |
| `DailyVolatility()` | F&O daily volatility report | `tradeDate` |
| `ExpiryDatesFuture()` | Upcoming futures expiry dates | - |
| `ExpiryDatesOptionIndex()` | Upcoming options expiry dates per index | - |
| `NSELiveOptionChain()` | Live option chain | `symbol`, `expiryDate` ("" for all), `oiMode` (`"full"`/`"compact"`) |
| `FIIDerivativesStatistics()` | FII derivatives stats | `tradeDate` |
| `FNOSecurityInBanPeriod()` | Securities in F&O ban | `tradeDate` |
| `LiveMostActiveUnderlying()` | Most active underlyings | - |
| `CategoryTurnoverFO()` | Category-wise turnover, F&O | `tradeDate` |
| `BusinessGrowthFOSegment()` | Business growth, F&O segment | `dataType`, `fromYear`, `toYear`, `month`, `year` |

**Instrument Types:**

| Code | Description |
|---|---|
| `FUTIDX` | Future Index |
| `FUTSTK` | Future Stock |
| `OPTIDX` | Option Index |
| `OPTSTK` | Option Stock |

**Examples:**

```go
// Futures price data
df, _ := derivatives.FuturePriceVolumeData("SBIN", "FUTSTK", "", "", "1M")

// Live option chain
df, _ = derivatives.NSELiveOptionChain("BANKNIFTY", "27-03-2025", "full")

// Compact option chain (fewer columns)
df, _ = derivatives.NSELiveOptionChain("NIFTY", "", "compact")

// FII derivatives statistics
df, _ = derivatives.FIIDerivativesStatistics("20-12-2025")
```

---

### Cash Market

```go
import "github.com/KartikeyaKotkar/nselib-go/cashmarket"
```

Pure facade over `mutualfunds` and `nsdlfpi`.

| Function | Description | Key Parameters |
|---|---|---|
| `NSDLFPIInvestmentActivity()` | NSDL FPI investment activity for a reporting date | `tradeDate` |
| `NSDLFPILatestInvestmentActivity()` | Latest NSDL FPI investment activity | - |
| `NSDLFPIDerivativeActivity()` | NSDL FPI derivative activity for a reporting date | `tradeDate` |
| `NSDLFPILatestDerivativeActivity()` | Latest NSDL FPI derivative activity | - |
| `AMFIMonthlyReportLinks()` | List AMFI monthly archive links | - |
| `AMFIMonthlyData()` | Parse one AMFI monthly report | `reportMonth`, `fileTypePriority...` |
| `AMFIMonthlyHistoricalData()` | Parse AMFI monthly reports across a range | `fromMonth`, `toMonth`, `fileTypePriority`, `includeAllVariants`, `strict` |

**Examples:**

```go
// NSDL FPI investment activity for a specific reporting date
df, _ := cashmarket.NSDLFPIInvestmentActivity("30-10-2025")

// Latest NSDL FPI derivative activity
df, _ = cashmarket.NSDLFPILatestDerivativeActivity()

// Parse a single AMFI monthly report
df, _ = cashmarket.AMFIMonthlyData("Jan-2026")

// Parse AMFI history for a month range
history, _ := cashmarket.AMFIMonthlyHistoricalData("Jan-2024", "Mar-2026", nil, false, false)
```

---

### Indices

```go
import "github.com/KartikeyaKotkar/nselib-go/indices"
```

| Function | Description | Key Parameters |
|---|---|---|
| `IndexList()` | Available indices by category | `indexCategory` |
| `ConstituentStockList()` | Stocks in a given index | `indexCategory`, `indexName` |
| `FactsheetURL()` | Factsheet PDF URL for an index | `indexCategory`, `indexName` |
| `LiveIndexPerformances()` | Live performance of all indices | - |

**Index Categories:** `BroadMarketIndices`, `SectoralIndices`, `ThematicIndices`, `StrategyIndices` (107 indices total)

**Examples:**

```go
// List all broad market indices
names, _ := indices.IndexList("BroadMarketIndices")

// Get Nifty 50 constituents
df, _ := indices.ConstituentStockList("BroadMarketIndices", "Nifty 50")

// Live index performances
df, _ = indices.LiveIndexPerformances()
```

---

### Debt

```go
import "github.com/KartikeyaKotkar/nselib-go/debt"
```

| Function | Description | Key Parameters |
|---|---|---|
| `SecuritiesAvailableForTrading()` | Debt securities available | `tradeDate` |

**Example:**

```go
df, _ := debt.SecuritiesAvailableForTrading("20-12-2025")
```

---

### Utilities

```go
import nselib "github.com/KartikeyaKotkar/nselib-go"
```

| Function | Description |
|---|---|
| `nselib.TradingHolidayCalendar()` | NSE trading holidays for all segments |
| `nselib.DeriveFromAndToDate()` | Resolve a `period` shorthand to a date range |
| `nselib.ParseCSV()` | CSV bytes to `DataFrame` with NSE column cleaning |
| `nselib.ExtractHTMLTables()` | HTML tables to header/row structs |

**Example:**

```go
df, _ := nselib.TradingHolidayCalendar()
```

---

## Logging & Debugging

The library is silent by default. Enable structured console logs via `log/slog`:

```go
import (
	"log/slog"

	nselib "github.com/KartikeyaKotkar/nselib-go"
	"github.com/KartikeyaKotkar/nselib-go/capitalmarket"
)

nselib.EnableLogging(slog.LevelDebug)

// Now calls emit trace logs
df, _ := capitalmarket.PriceVolumeData("SBIN", "", "", "1W")
```

## Performance

Cookies prime once per origin (Python primes before every call) and history windows, XBRL filings,
AMFI months, and index expiries fetch concurrently with bounded parallelism. Measured on
NIFTY OPTIDX CE 1Y (208,906 rows): Go 7.8s vs Python 20.9s.

## Notes

- `nsdlfpi` browser automation needs a Chrome/Chromium executable; without one it uses the plain-HTTP fallback.
- AMFI PDF parsing is text-only (no `pdfplumber`-style table extraction); Excel/HTML reports keep full structure.
- Legacy `.xls` files parse via `github.com/extrame/xls`; `.xlsx` via `excelize`.
- NSE decides who it serves; some regions see 403s on certain endpoints. The client re-primes and surfaces typed `NSEError` values.

## Development

```bash
go build ./...
go vet ./...
go test ./... -race
```

## How to Contribute

### Report Issues & Suggest Features

Found a bug or have a feature request? Open an issue on the [GitHub Issues page](https://github.com/KartikeyaKotkar/nselib-go/issues).

### Submit Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -m 'Add your feature'`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Open a Pull Request

### Contact

- **Go port:** [KartikeyaKotkar/nselib-go](https://github.com/KartikeyaKotkar/nselib-go)
- **Original Python library:** [RuchiTanmay/nselib](https://github.com/RuchiTanmay/nselib) by [Ruchi Tanmay](https://www.linkedin.com/in/ruchi-tanmay-61848219)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
