// Package cashmarket ports nselib/cash_market to Go.
// Pure facade: delegates to mutualfunds and nsdlfpi like the Python version.
package cashmarket

import (
	"github.com/KartikeyaKotkar/nselib-go"
	"github.com/KartikeyaKotkar/nselib-go/mutualfunds"
	"github.com/KartikeyaKotkar/nselib-go/nsdlfpi"
)

// NSDLFPIInvestmentActivity returns NSDL FPI investment activity for a trade date.
func NSDLFPIInvestmentActivity(tradeDate string) (nselib.DataFrame, error) {
	return nsdlfpi.FetchInvestmentActivity(tradeDate)
}

// NSDLFPILatestInvestmentActivity returns the latest NSDL FPI investment activity.
func NSDLFPILatestInvestmentActivity() (nselib.DataFrame, error) {
	return nsdlfpi.FetchLatestInvestmentActivity()
}

// NSDLFPIDerivativeActivity returns NSDL FPI derivative activity for a trade date.
func NSDLFPIDerivativeActivity(tradeDate string) (nselib.DataFrame, error) {
	return nsdlfpi.FetchDerivativeActivity(tradeDate)
}

// NSDLFPILatestDerivativeActivity returns the latest NSDL FPI derivative activity.
func NSDLFPILatestDerivativeActivity() (nselib.DataFrame, error) {
	return nsdlfpi.FetchLatestDerivativeActivity()
}

// AMFIMonthlyReportLinks lists all AMFI monthly report links.
func AMFIMonthlyReportLinks() (nselib.DataFrame, error) {
	return mutualfunds.AMFIMonthlyReportLinks()
}

// AMFIMonthlyData fetches a single month AMFI report.
func AMFIMonthlyData(reportMonth string, fileTypePriority ...string) (nselib.DataFrame, error) {
	return mutualfunds.AMFIMonthlyData(reportMonth, fileTypePriority...)
}

// AMFIMonthlyHistoricalData fetches AMFI reports across a month range.
func AMFIMonthlyHistoricalData(fromMonth, toMonth string, fileTypePriority []string, includeAllVariants, strict bool) (nselib.DataFrame, error) {
	return mutualfunds.AMFIMonthlyHistoricalData(fromMonth, toMonth, fileTypePriority, includeAllVariants, strict)
}
