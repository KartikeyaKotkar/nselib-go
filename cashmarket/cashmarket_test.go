package cashmarket

import (
	"testing"

	"github.com/KartikeyaKotkar/nselib-go"
)

// Compile-time delegate signature checks (no network): facade must match
// the underlying package function shapes.
var (
	_ func(string) (nselib.DataFrame, error)            = NSDLFPIInvestmentActivity
	_ func() (nselib.DataFrame, error)                  = NSDLFPILatestInvestmentActivity
	_ func(string) (nselib.DataFrame, error)            = NSDLFPIDerivativeActivity
	_ func() (nselib.DataFrame, error)                  = NSDLFPILatestDerivativeActivity
	_ func() (nselib.DataFrame, error)                  = AMFIMonthlyReportLinks
	_ func(string, ...string) (nselib.DataFrame, error) = AMFIMonthlyData
)

func TestFacadeSignatures(t *testing.T) {
	// Existence enforced at compile time above; nothing to fetch offline.
}
