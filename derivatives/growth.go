package derivatives

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

// NormalizeBusinessGrowthFinancialYear splits "2025-2026" style input.
func NormalizeBusinessGrowthFinancialYear(fromYear, toYear string) (string, string, error) {
	if toYear == "" && fromYear != "" && strings.Contains(fromYear, "-") {
		parts := strings.SplitN(fromYear, "-", 2)
		fromYear, toYear = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	if fromYear == "" || toYear == "" {
		return "", "", nselib.NewDataNotFoundError("For monthly data provide from_year and to_year, e.g. from_year='2025', to_year='2026'")
	}
	return strings.TrimSpace(fromYear), strings.TrimSpace(toYear), nil
}

// NormalizeBusinessGrowthDailyArgs normalizes month/year ("Mar-26" split, 2-digit year expanded).
func NormalizeBusinessGrowthDailyArgs(month, year string) (string, string, error) {
	if year == "" && month != "" && strings.Contains(month, "-") {
		parts := strings.SplitN(month, "-", 2)
		month, year = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	if month == "" || year == "" {
		return "", "", nselib.NewDataNotFoundError("For daily data provide month and year, e.g. month='Mar', year='2026'")
	}
	month = strings.TrimSpace(month)
	if len(month) >= 3 {
		if t, err := time.Parse("Jan", title3(month[:3])); err == nil {
			month = t.Format("Jan")
		} else {
			return "", "", nselib.NewDataNotFoundError("Month should be a valid month name like 'Mar' or 'March'")
		}
	}
	year = strings.TrimSpace(year)
	if len(year) == 2 {
		if _, err := strconv.Atoi(year); err == nil {
			year = "20" + year
		}
	}
	return month, year, nil
}

func title3(s string) string {
	s = strings.ToLower(s)
	return strings.ToUpper(s[:1]) + s[1:]
}

// businessGrowthDataFrame flattens {"data": [{"data": {...}}, ...]} + numeric coercion.
func businessGrowthDataFrame(raw map[string]interface{}) nselib.DataFrame {
	arr, _ := raw["data"].([]interface{})
	df := make(nselib.DataFrame, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if inner, ok := m["data"].(map[string]interface{}); ok {
			df = append(df, nselib.Record(inner))
		} else {
			df = append(df, nselib.Record(m))
		}
	}
	if len(df) == 0 {
		return df
	}
	preserve := map[string]bool{"type": true, "TYPE": true, "date": true}
	cols := map[string]bool{}
	for _, r := range df {
		for k := range r {
			cols[k] = true
		}
	}
	for c := range cols {
		if preserve[c] {
			continue
		}
		allNum := true
		for _, r := range df {
			s, isStr := r[c].(string)
			if !isStr {
				if _, ok := r[c].(float64); !ok {
					allNum = false
					break
				}
				continue
			}
			if s == "" || s == "-" || s == "--" {
				continue
			}
			if _, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(s), ",", ""), 64); err != nil {
				allNum = false
				break
			}
		}
		if allNum {
			for _, r := range df {
				if s, ok := r[c].(string); ok && s != "" && s != "-" && s != "--" {
					if f, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(s), ",", ""), 64); err == nil {
						r[c] = f
					}
				}
			}
		}
	}
	return df
}

// BusinessGrowthFOSegment fetches yearly/monthly/daily F&O business growth.
func BusinessGrowthFOSegment(dataType, fromYear, toYear, month, year string) (nselib.DataFrame, error) {
	if err := nselib.ValidateParamFromList(dataType, []string{"yearly", "monthly", "daily"}); err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	var err error
	switch dataType {
	case "yearly":
		raw, err = getBusinessGrowthRaw("/api/historicalOR/fo/tbg/yearly")
	case "monthly":
		var fy, ty string
		if fy, ty, err = NormalizeBusinessGrowthFinancialYear(fromYear, toYear); err != nil {
			return nil, err
		} else {
			raw, err = getBusinessGrowthRaw(fmt.Sprintf("/api/historicalOR/fo/tbg/monthly?from=%s&to=%s", fy, ty))
		}
	default:
		var m, y string
		if m, y, err = NormalizeBusinessGrowthDailyArgs(month, year); err != nil {
			return nil, err
		} else {
			raw, err = getBusinessGrowthRaw(fmt.Sprintf("/api/historicalOR/fo/tbg/daily?month=%s&year=%s", m, y))
		}
	}
	if err != nil {
		return nil, err
	}
	return businessGrowthDataFrame(raw), nil
}
