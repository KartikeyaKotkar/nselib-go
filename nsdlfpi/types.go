// Package nsdlfpi ports nselib/nsdl_fpi to Go.
package nsdlfpi

import (
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

// Report date layout dd-Mon-YYYY.
const ReportDateLayout = "02-Jan-2006"

// NSDL endpoints with pilot-first failover like Python's REQUEST_BASE_URLS.
const (
	ProductionBaseURL = "https://www.fpi.nsdl.co.in/web/Reports"
	PilotBaseURL      = "https://pilot.fpi.nsdl.co.in/Reports"
	LatestPage        = "Latest.aspx"
	ArchivePage       = "Archive.aspx"
)

// ArchiveHiddenFields are the ASP.NET fields posted for archive queries.
var ArchiveHiddenFields = []string{"__VIEWSTATE", "__VIEWSTATEGENERATOR", "__EVENTVALIDATION"}

// InvestmentColumns mirrors Python's INVESTMENT_COLUMNS.
var InvestmentColumns = []string{
	"REPORT_DATE", "ASSET_CLASS", "INVESTMENT_ROUTE",
	"GROSS_PURCHASES_RS_CR", "GROSS_SALES_RS_CR",
	"NET_INVESTMENT_RS_CR", "NET_INVESTMENT_USD_MN", "USD_INR_CONVERSION",
}

// DerivativeColumns mirrors Python's DERIVATIVE_COLUMNS.
var DerivativeColumns = []string{
	"REPORT_DATE", "DERIVATIVE_PRODUCT",
	"BUY_CONTRACTS", "BUY_AMOUNT_CR", "SELL_CONTRACTS", "SELL_AMOUNT_CR",
	"OPEN_INTEREST_CONTRACTS", "OPEN_INTEREST_AMOUNT_CR",
}

// DailyDerivativeProducts is the allow-list for derivative rows.
var DailyDerivativeProducts = map[string]bool{
	"Index Futures": true, "Index Options": true,
	"Stock Futures": true, "Stock Options": true,
	"Interest Rate Futures": true, "Currency Futures": true,
	"Currency Options": true, "Commodity Futures": true,
	"Commodity Options": true,
}

// ReportBundle mirrors NSDLFPIReportBundle.
type ReportBundle struct {
	Investment nselib.DataFrame
	Derivative nselib.DataFrame
	SourcePage string
	AsOfDate   *time.Time
}

// MaxBundleReportDate returns the latest REPORT_DATE across both frames.
func MaxBundleReportDate(b *ReportBundle) *time.Time {
	var best *time.Time
	for _, df := range []nselib.DataFrame{b.Investment, b.Derivative} {
		for _, r := range df {
			s, _ := r["REPORT_DATE"].(string)
			if s == "" {
				continue
			}
			if t, err := time.Parse(ReportDateLayout, s); err == nil {
				if best == nil || t.After(*best) {
					cp := t
					best = &cp
				}
			}
		}
	}
	return best
}
