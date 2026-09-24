package capitalmarket

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

// xbrlKeys mirrors keys_to_extract in get_func.py.
var xbrlKeys = []string{
	"ScripCode", "Symbol", "MSEISymbol", "NameOfTheCompany", "ClassOfSecurity",
	"DateOfStartOfFinancialYear", "DateOfEndOfFinancialYear",
	"DateOfBoardMeetingWhenFinancialResultsWereApproved",
	"DateOnWhichPriorIntimationOfTheMeetingForConsideringFinancialResultsWasInformedToTheExchange",
	"DescriptionOfPresentationCurrency", "LevelOfRoundingUsedInFinancialStatements",
	"ReportingQuarter", "StartTimeOfBoardMeeting", "EndTimeOfBoardMeeting",
	"DateOfStartOfBoardMeeting", "DateOfEndOfBoardMeeting",
	"DeclarationOfUnmodifiedOpinionOrStatementOnImpactOfAuditQualification",
	"IsCompanyReportingMultisegmentOrSingleSegment", "DescriptionOfSingleSegment",
	"DateOfStartOfReportingPeriod", "DateOfEndOfReportingPeriod",
	"WhetherResultsAreAuditedOrUnaudited", "NatureOfReportStandaloneConsolidated",
	"RevenueFromOperations", "OtherIncome", "Income", "CostOfMaterialsConsumed",
	"PurchasesOfStockInTrade", "ChangesInInventoriesOfFinishedGoodsWorkInProgressAndStockInTrade",
	"EmployeeBenefitExpense", "FinanceCosts", "DepreciationDepletionAndAmortisationExpense",
	"OtherExpenses", "Expenses", "ProfitBeforeExceptionalItemsAndTax",
	"ExceptionalItemsBeforeTax", "ProfitBeforeTax", "CurrentTax", "DeferredTax",
	"TaxExpense", "NetMovementInRegulatoryDeferralAccountBalancesRelatedToProfitOrLossAndTheRelatedDeferredTaxMovement",
	"ProfitLossForPeriodFromContinuingOperations", "ProfitLossFromDiscontinuedOperationsBeforeTax",
	"TaxExpenseOfDiscontinuedOperations", "ProfitLossFromDiscontinuedOperationsAfterTax",
	"ShareOfProfitLossOfAssociatesAndJointVenturesAccountedForUsingEquityMethod",
	"ProfitLossForPeriod", "OtherComprehensiveIncomeNetOfTaxes", "ComprehensiveIncomeForThePeriod",
	"ProfitOrLossAttributableToOwnersOfParent", "ProfitOrLossAttributableToNonControllingInterests",
	"ComprehensiveIncomeForThePeriodAttributableToOwnersOfParent",
	"ComprehensiveIncomeForThePeriodAttributableToOwnersOfParentNonControllingInterests",
	"PaidUpValueOfEquityShareCapital", "FaceValueOfEquityShareCapital",
	"BasicEarningsLossPerShareFromContinuingOperations", "DilutedEarningsLossPerShareFromContinuingOperations",
	"BasicEarningsLossPerShareFromDiscontinuedOperations", "DilutedEarningsLossPerShareFromDiscontinuedOperations",
	"BasicEarningsLossPerShareFromContinuingAndDiscontinuedOperations",
	"DilutedEarningsLossPerShareFromContinuingAndDiscontinuedOperations",
	"DescriptionOfOtherExpenses", "DescriptionOfItemThatWillNotBeReclassifiedToProfitAndLoss",
	"AmountOfItemThatWillNotBeReclassifiedToProfitAndLoss",
	"IncomeTaxRelatingToItemsThatWillNotBeReclassifiedToProfitOrLoss",
	"DescriptionOfItemThatWillBeReclassifiedToProfitAndLoss",
	"AmountOfItemThatWillBeReclassifiedToProfitAndLoss",
	"IncomeTaxRelatingToItemsThatWillBeReclassifiedToProfitOrLoss",
}

// extractXBRLValues pulls in-bse-fin namespaced values via regex (stdlib only).
func extractXBRLValues(xmlBody []byte, keys []string) nselib.Record {
	s := string(xmlBody)
	rec := nselib.Record{}
	for _, k := range keys {
		// matches <in-bse-fin:KEY ...>value</...:KEY> and <KEY>value</KEY>
		re := regexp.MustCompile(`(?s)<(?:[A-Za-z0-9_.-]+:)?` + regexp.QuoteMeta(k) + `(?:\s[^>]*)?>(.*?)</(?:[A-Za-z0-9_.-]+:)?` + regexp.QuoteMeta(k) + `>`)
		if m := re.FindStringSubmatch(s); m != nil {
			rec[k] = strings.TrimSpace(m[1])
		} else {
			rec[k] = nil
		}
	}
	return rec
}

var xbrlHTTP = &http.Client{Timeout: 30 * time.Second}

// FinancialResultsForEquity fetches XBRL financial results per master row.
func FinancialResultsForEquity(fromDate, toDate, period string, foSec bool, finPeriod string) (nselib.DataFrame, error) {
	if finPeriod == "" {
		finPeriod = "Quarterly"
	}
	from, to, err := resolveDateRange(fromDate, toDate, period)
	if err != nil {
		return nil, err
	}
	master, err := getFinancialMasterRaw(
		from.Format(nselib.LayoutDDMMYYYY), to.Format(nselib.LayoutDDMMYYYY), foSec, finPeriod)
	if err != nil {
		return nil, err
	}
	var out nselib.DataFrame
	for _, row := range master {
		xbrlURL, _ := row["xbrl"].(string)
		if xbrlURL == "" {
			// fall back to case-insensitive lookup
			for k, v := range row {
				if strings.EqualFold(k, "xbrl") {
					xbrlURL, _ = v.(string)
					break
				}
			}
		}
		if xbrlURL == "" {
			continue
		}
		req, err := http.NewRequest("GET", xbrlURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
		req.Header.Set("Referer", "https://www.nseindia.com/")
		resp, err := xbrlHTTP.Do(req)
		if err != nil {
			return nil, nselib.NewAPIError(fmt.Sprintf("fetch XBRL: %v", err))
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, nselib.NewAPIError(fmt.Sprintf("fetch XBRL: status %d", resp.StatusCode))
		}
		out = append(out, extractXBRLValues(body, xbrlKeys))
	}
	if out == nil {
		out = nselib.DataFrame{}
	}
	return out, nil
}
