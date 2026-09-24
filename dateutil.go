package nselib

import (
	"fmt"
	"strings"
	"time"
)

// ValidateParamFromList mirrors validate_param_from_list in libutil.py.
func ValidateParamFromList(value string, options []string) error {
	for _, o := range options {
		if value == o {
			return nil
		}
	}
	return fmt.Errorf("'%s' not a valid parameter :: select from %v", value, options)
}

// ValidateDateParams validates from_date/to_date/period parameters.
func ValidateDateParams(fromDate, toDate, period string) error {
	if period == "" && (fromDate == "" || toDate == "") {
		return fmt.Errorf("please provide the valid parameters")
	}
	if period != "" {
		if err := ValidateParamFromList(strings.ToUpper(period), EquityPeriods); err != nil {
			return fmt.Errorf("period = %s is not a valid value", period)
		}
		return nil
	}
	from, err1 := time.Parse(LayoutDDMMYYYY, fromDate)
	to, err2 := time.Parse(LayoutDDMMYYYY, toDate)
	if err1 != nil || err2 != nil {
		return fmt.Errorf("either or both from_date = %s || to_date = %s are not valid value", fromDate, toDate)
	}
	if int(to.Sub(from).Hours()/24) < 1 {
		return fmt.Errorf("to_date should greater than from_date")
	}
	return nil
}

// SubtractMonths accurately subtracts months handling leap years.
func SubtractMonths(t time.Time, months int) time.Time {
	year, month := t.Year(), int(t.Month())-months
	for month <= 0 {
		month += 12
		year--
	}
	daysIn := daysInMonth(year, time.Month(month))
	day := t.Day()
	if day > daysIn {
		day = daysIn
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, t.Location())
}

func daysInMonth(year int, month time.Month) int {
	switch month {
	case time.February:
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
			return 29
		}
		return 28
	case time.April, time.June, time.September, time.November:
		return 30
	default:
		return 31
	}
}

// DeriveFromAndToDate computes date range from a period string.
// Uses NSE trading calendar to find valid trading days.
// Fixes Python bug: 3M now handled via SubtractMonths(today, 3).
func DeriveFromAndToDate(fromDate, toDate, period string) (string, string, error) {
	if period == "" {
		return fromDate, toDate, nil
	}
	today := time.Now()
	var fDate time.Time
	switch strings.ToUpper(period) {
	case "1D":
		fDate = today.AddDate(0, 0, -1)
	case "1W":
		fDate = today.AddDate(0, 0, -7)
	case "1M":
		fDate = SubtractMonths(today, 1)
	case "3M":
		fDate = SubtractMonths(today, 3)
	case "6M":
		fDate = SubtractMonths(today, 6)
	case "1Y":
		fDate = SubtractMonths(today, 12)
	default:
		return "", "", fmt.Errorf("period = %s is not a valid value", period)
	}
	for {
		ok, err := IsNSETradingDay(fDate)
		if err != nil {
			break // calendar unavailable: keep derived date
		}
		if ok {
			break
		}
		fDate = fDate.AddDate(0, 0, -1)
	}
	return fDate.Format(LayoutDDMMYYYY), today.Format(LayoutDDMMYYYY), nil
}

// CleanNSESymbol URL-encodes & and uppercases the symbol.
func CleanNSESymbol(symbol string) string {
	return strings.ToUpper(strings.ReplaceAll(symbol, "&", "%26"))
}

// GetMonthFromDate returns Mon abbrev for YYYY-mm-dd input.
func GetMonthFromDate(tradeDate string) (string, error) {
	t, err := time.Parse(LayoutYYYYMMDD, tradeDate)
	if err != nil {
		return "", fmt.Errorf("trade_date = %s not valid, want YYYY-mm-dd", tradeDate)
	}
	return t.Format(LayoutMon), nil
}
