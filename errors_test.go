package nselib

import (
	"strings"
	"testing"
)

func TestErrorCodes(t *testing.T) {
	cases := []struct {
		err  *NSEError
		code ErrorCode
	}{
		{NewAPIError("x"), ErrCodeAPI},
		{NewDataNotFoundError("x"), ErrCodeDataNotFound},
		{NewIndexDataNotFoundError("x"), ErrCodeIndexDataNotFound},
		{NewCalendarNotFoundError("x"), ErrCodeCalendarNotFound},
		{NewInvalidIndexCategoryError("x"), ErrCodeInvalidIndexCategory},
		{NewInvalidIndexError("x"), ErrCodeInvalidIndex},
		{NewDerivativeInstrumentNotFoundError("x"), ErrCodeDerivativeInstrumentNotFound},
	}
	for _, c := range cases {
		if c.err.Code != c.code {
			t.Errorf("want %s got %s", c.code, c.err.Code)
		}
		if !strings.Contains(c.err.Error(), string(c.code)) {
			t.Errorf("Error() missing code: %s", c.err.Error())
		}
	}
}

func TestIndexDataNotFoundDistinct(t *testing.T) {
	a := NewDataNotFoundError("m")
	b := NewIndexDataNotFoundError("m")
	if a.Code != b.Code {
		t.Fatalf("codes should match DATA_NOT_FOUND, got %s vs %s", a.Code, b.Code)
	}
}
