package nselib

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	holidayCache map[string]bool
	holidayOnce  sync.Once
	holidayErr   error
	holMu        sync.RWMutex
)

// productNames maps NSE API product codes to display names (mirrors libutil.py).
var productNames = map[string]string{
	"CBM":  "Corporate Bonds",
	"CD":   "Currency Derivatives",
	"CM":   "Equities",
	"CMOT": "CMOT",
	"COM":  "Commodity Derivatives",
	"FO":   "Equity Derivatives",
	"IRD":  "Interest Rate Derivatives",
	"MF":   "Mutual Funds",
	"NDM":  "New Debt Segment",
	"NTRP": "Negotiated Trade Reporting Platform",
	"SLBS": "Securities Lending & Borrowing Schemes",
}

// holidayMasterResponse mirrors NSE holiday-master JSON: product -> list of entries.
type holidayEntry struct {
	TradingDate string `json:"tradingDate"`
	Weekday     string `json:"weekDay"`
	Description string `json:"description"`
	SrNo        int    `json:"Sr_no"`
}

func loadHolidays() {
	holidayOnce.Do(func() {
		m := map[string]bool{}
		client := NewNSEClient()
		var raw map[string][]holidayEntry
		if err := client.FetchJSON(
			"https://www.nseindia.com/api/holiday-master?type=trading",
			DefaultOriginURL, &raw,
		); err != nil {
			holidayErr = err
			return
		}
		for _, entries := range raw {
			for _, e := range entries {
				m[strings.TrimSpace(e.TradingDate)] = true
			}
		}
		holMu.Lock()
		holidayCache = m
		holMu.Unlock()
	})
}

// IsNSETradingDay checks if a date is a valid NSE trading day.
// Mon-Fri and not in holiday set. Falls back to weekday-only if calendar fetch fails.
func IsNSETradingDay(t time.Time) (bool, error) {
	if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		return false, nil
	}
	loadHolidays()
	if holidayErr != nil {
		return true, holidayErr
	}
	holMu.RLock()
	defer holMu.RUnlock()
	key1 := t.Format(LayoutDDMMYYYY)
	key2 := t.Format("02-Jan-2006")
	if holidayCache[key1] || holidayCache[key2] {
		return false, nil
	}
	return true, nil
}

// TradingHolidayCalendar fetches and returns the NSE holiday calendar.
func TradingHolidayCalendar() (DataFrame, error) {
	client := NewNSEClient()
	var raw map[string][]holidayEntry
	if err := client.FetchJSON(
		"https://www.nseindia.com/api/holiday-master?type=trading",
		DefaultOriginURL, &raw,
	); err != nil {
		return nil, NewCalendarNotFoundError(fmt.Sprintf("calendar data not found: %v", err))
	}
	var df DataFrame
	for prod, entries := range raw {
		name := productNames[prod]
		if name == "" {
			name = "Unknown"
		}
		for _, e := range entries {
			df = append(df, Record{
				"Product":     name,
				"tradingDate": e.TradingDate,
				"weekDay":     e.Weekday,
				"description": e.Description,
				"Sr_no":       e.SrNo,
			})
		}
	}
	if len(df) == 0 {
		return nil, NewCalendarNotFoundError("calendar data not found try after some time")
	}
	return df, nil
}
