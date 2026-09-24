package nselib

import (
	"testing"
	"time"
)

func TestValidateDateParams(t *testing.T) {
	if err := ValidateDateParams("01-01-2024", "05-01-2024", ""); err != nil {
		t.Errorf("valid dates rejected: %v", err)
	}
	if err := ValidateDateParams("", "", ""); err == nil {
		t.Error("empty params should fail")
	}
	if err := ValidateDateParams("", "", "9M"); err == nil {
		t.Error("bad period should fail")
	}
	if err := ValidateDateParams("05-01-2024", "01-01-2024", ""); err == nil {
		t.Error("reversed dates should fail")
	}
	if err := ValidateDateParams("", "", "3M"); err != nil {
		t.Errorf("3M should be valid: %v", err)
	}
}

func TestSubtractMonths(t *testing.T) {
	// leap year edge: 31 Mar 2024 - 1 month = 29 Feb 2024
	d := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)
	got := SubtractMonths(d, 1)
	if got.Day() != 29 || got.Month() != time.February {
		t.Errorf("leap subtract failed: %v", got)
	}
	// year wrap: Jan 2024 - 1 = Dec 2023
	d2 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	got2 := SubtractMonths(d2, 1)
	if got2.Year() != 2023 || got2.Month() != time.December {
		t.Errorf("year wrap failed: %v", got2)
	}
}

func TestDeriveFromAndToDate3M(t *testing.T) {
	from, to, err := DeriveFromAndToDate("", "", "3M")
	if err != nil {
		t.Fatal(err)
	}
	if from == "" || to == "" {
		t.Error("empty derived dates")
	}
	f, err1 := time.Parse(LayoutDDMMYYYY, from)
	td, err2 := time.Parse(LayoutDDMMYYYY, to)
	if err1 != nil || err2 != nil {
		t.Fatalf("bad format: %v %v", err1, err2)
	}
	if !f.Before(td) {
		t.Errorf("from %s not before to %s", from, to)
	}
}

func TestCleanNSESymbol(t *testing.T) {
	if got := CleanNSESymbol("m&m"); got != "M%26M" {
		t.Errorf("got %s", got)
	}
}

func TestGetMonthFromDate(t *testing.T) {
	m, err := GetMonthFromDate("2024-12-25")
	if err != nil || m != "Dec" {
		t.Errorf("got %s %v", m, err)
	}
}
