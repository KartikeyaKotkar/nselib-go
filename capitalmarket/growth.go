package capitalmarket

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
		return "", "", fmt.Errorf("for monthly data provide from_year and to_year, e.g. from_year='2025', to_year='2026'")
	}
	return strings.TrimSpace(fromYear), strings.TrimSpace(toYear), nil
}

// NormalizeBusinessGrowthDailyArgs normalizes month/year ("Mar-26" split, 4-digit year trim).
func NormalizeBusinessGrowthDailyArgs(month, year string) (string, string, error) {
	if year == "" && month != "" && strings.Contains(month, "-") {
		parts := strings.SplitN(month, "-", 2)
		month, year = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	if month == "" || year == "" {
		return "", "", fmt.Errorf("for daily data provide month and year, e.g. month='Mar', year='26'")
	}
	month = strings.TrimSpace(month)
	if len(month) >= 3 {
		if t, err := time.Parse("Jan", title3(month[:3])); err == nil {
			month = t.Format("Jan")
		} else {
			return "", "", fmt.Errorf("month should be a valid month name like 'Mar' or 'March'")
		}
	}
	year = strings.TrimSpace(year)
	if len(year) == 4 {
		if _, err := strconv.Atoi(year); err == nil {
			year = year[2:]
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
	// light numeric coercion: strip commas, coerce pure-numeric columns
	if len(df) == 0 {
		return df
	}
	cols := map[string]bool{}
	for _, r := range df {
		for k := range r {
			cols[k] = true
		}
	}
	for c := range cols {
		if c == "type" || c == "TYPE" || c == "GLY_MONTH_YEAR" || c == "GLM_MONTH_YEAR" || c == "F_TIMESTAMP" {
			continue
		}
		allNum := true
		for _, r := range df {
			s, _ := r[c].(string)
			if s == "" || s == "-" || s == "--" {
				continue
			}
			sc := strings.ReplaceAll(s, ",", "")
			if _, err := strconv.ParseFloat(strings.TrimSpace(sc), 64); err != nil {
				if _, ok := r[c].(float64); !ok {
					allNum = false
					break
				}
			}
		}
		if allNum {
			for _, r := range df {
				if s, ok := r[c].(string); ok {
					if f, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(s), ",", ""), 64); err == nil {
						r[c] = f
					}
				}
			}
		}
	}
	return df
}

// BusinessGrowthCMSegment fetches yearly/monthly/daily CM business growth.
func BusinessGrowthCMSegment(dataType, fromYear, toYear, month, year string) (nselib.DataFrame, error) {
	if err := nselib.ValidateParamFromList(dataType, []string{"yearly", "monthly", "daily"}); err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	var err error
	switch dataType {
	case "yearly":
		raw, err = getBusinessGrowthRaw("/api/historicalOR/cm/tbg/yearly")
	case "monthly":
		var fy, ty string
		if fy, ty, err = NormalizeBusinessGrowthFinancialYear(fromYear, toYear); err != nil {
			return nil, err
		} else {
			raw, err = getBusinessGrowthRaw(fmt.Sprintf("/api/historicalOR/cm/tbg/monthly?from=%s&to=%s", fy, ty))
		}
	default:
		var m, y string
		if m, y, err = NormalizeBusinessGrowthDailyArgs(month, year); err != nil {
			return nil, err
		} else {
			raw, err = getBusinessGrowthRaw(fmt.Sprintf("/api/historicalOR/cm/tbg/daily?month=%s&year=%s", m, y))
		}
	}
	if err != nil {
		return nil, err
	}
	return businessGrowthDataFrame(raw), nil
}
