package nselib

import (
	"strings"
	"testing"
)

func TestParseCSVBasic(t *testing.T) {
	in := "Symbol, Series ,FH_Date\nSBIN,EQ,01-01-2024\n"
	df, err := ParseCSV(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(df) != 1 {
		t.Fatalf("want 1 row got %d", len(df))
	}
	if _, ok := df[0]["Date"]; !ok {
		t.Errorf("FH_ prefix not stripped: %v", df[0])
	}
	if df[0]["Symbol"] != "SBIN" {
		t.Errorf("bad value: %v", df[0])
	}
}

func TestParseCSVSkipRows(t *testing.T) {
	in := "junk\nSymbol,Series\nSBIN,EQ\n"
	df, err := ParseCSV(strings.NewReader(in), SkipRows(1))
	if err != nil {
		t.Fatal(err)
	}
	if len(df) != 1 || df[0]["Symbol"] != "SBIN" {
		t.Errorf("skip rows failed: %v", df)
	}
}

func TestParseCSVEmpty(t *testing.T) {
	_, err := ParseCSV(strings.NewReader(""))
	if err == nil {
		t.Error("want error on empty")
	}
}

func TestCleanHelpers(t *testing.T) {
	got := StripColumnPrefixes([]string{"FH_DATE", "EOD_OPEN", "HIT_X"})
	if got[0] != "DATE" || got[1] != "OPEN" || got[2] != "X" {
		t.Errorf("strip failed: %v", got)
	}
}
