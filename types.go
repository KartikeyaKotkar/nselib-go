package nselib

// Record represents a single row of data (like a dict in Python).
type Record = map[string]interface{}

// DataFrame represents tabular data as a slice of records.
// Each Record is a row with column names as keys.
// Go equivalent of pandas.DataFrame used throughout the library.
type DataFrame = []Record
