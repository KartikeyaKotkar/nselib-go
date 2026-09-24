package nselib

import (
	"encoding/csv"
	"io"
	"strings"
)

// CSVOption configures ParseCSV.
type CSVOption func(*csvOptions)

type csvOptions struct {
	skipRows     int
	cleanColumns bool
}

// SkipRows skips first N rows before header.
func SkipRows(n int) CSVOption {
	return func(o *csvOptions) { o.skipRows = n }
}

// WithoutColumnCleaning disables header cleaning.
func WithoutColumnCleaning() CSVOption {
	return func(o *csvOptions) { o.cleanColumns = false }
}

// unwantedPrefixes mirrors cleaning_column_name in libutil.py.
var unwantedPrefixes = []string{"FH_", "EOD_", "HIT_"}

// CleanColumnNames strips spaces from column names.
func CleanColumnNames(columns []string) []string {
	out := make([]string, len(columns))
	for i, c := range columns {
		out[i] = strings.TrimSpace(c)
	}
	return out
}

// StripColumnPrefixes removes FH_, EOD_, HIT_ prefixes.
func StripColumnPrefixes(columns []string) []string {
	out := make([]string, len(columns))
	for i, c := range columns {
		for _, p := range unwantedPrefixes {
			c = strings.ReplaceAll(c, p, "")
		}
		out[i] = c
	}
	return out
}

// ParseCSV reads CSV data and returns a DataFrame.
// Options control skipRows and column cleaning.
func ParseCSV(r io.Reader, opts ...CSVOption) (DataFrame, error) {
	o := csvOptions{cleanColumns: true}
	for _, opt := range opts {
		opt(&o)
	}
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, NewDataNotFoundError("parse CSV: " + err.Error())
	}
	for i := 0; i < o.skipRows && len(records) > 0; i++ {
		records = records[1:]
	}
	if len(records) < 1 {
		return nil, NewDataNotFoundError("parse CSV: empty data")
	}
	header := records[0]
	if o.cleanColumns {
		header = CleanColumnNames(header)
		header = StripColumnPrefixes(header)
	}
	var df DataFrame
	for _, row := range records[1:] {
		if len(row) == 0 {
			continue
		}
		rec := make(Record, len(header))
		for i, h := range header {
			if i < len(row) {
				rec[h] = strings.TrimSpace(row[i])
			} else {
				rec[h] = ""
			}
		}
		df = append(df, rec)
	}
	return df, nil
}
