package nselib

import "fmt"

// ErrorCode represents categorized error codes for NSElib.
type ErrorCode string

const (
	ErrCodeUnknown                      ErrorCode = "UNKNOWN_ERROR"
	ErrCodeAPI                          ErrorCode = "API_ERROR"
	ErrCodeCalendarNotFound             ErrorCode = "CALENDAR_NOT_FOUND"
	ErrCodeDataNotFound                 ErrorCode = "DATA_NOT_FOUND"
	ErrCodeIndexDataNotFound            ErrorCode = "DATA_NOT_FOUND" // Same code as DataNotFound, distinct constructor for API parity
	ErrCodeInvalidIndexCategory         ErrorCode = "INVALID_INDEX_CATEGORY"
	ErrCodeInvalidIndex                 ErrorCode = "INVALID_INDEX"
	ErrCodeDerivativeInstrumentNotFound ErrorCode = "DERIVATIVE_INSTRUMENT_NOT_FOUND"
)

// NSEError is the base error type for all nselib errors.
type NSEError struct {
	Code    ErrorCode
	Message string
}

// Error returns the formatted "[CODE] message" string.
func (e *NSEError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAPIError creates an API error.
func NewAPIError(msg string) *NSEError { return &NSEError{Code: ErrCodeAPI, Message: msg} }

// NewDataNotFoundError creates a data-not-found error.
func NewDataNotFoundError(msg string) *NSEError {
	return &NSEError{Code: ErrCodeDataNotFound, Message: msg}
}

// NewIndexDataNotFoundError creates an index-data-not-found error (API parity).
func NewIndexDataNotFoundError(msg string) *NSEError {
	return &NSEError{Code: ErrCodeIndexDataNotFound, Message: msg}
}

// NewCalendarNotFoundError creates a calendar-not-found error.
func NewCalendarNotFoundError(msg string) *NSEError {
	return &NSEError{Code: ErrCodeCalendarNotFound, Message: msg}
}

// NewInvalidIndexCategoryError creates an invalid-index-category error.
func NewInvalidIndexCategoryError(msg string) *NSEError {
	return &NSEError{Code: ErrCodeInvalidIndexCategory, Message: msg}
}

// NewInvalidIndexError creates an invalid-index error.
func NewInvalidIndexError(msg string) *NSEError {
	return &NSEError{Code: ErrCodeInvalidIndex, Message: msg}
}

// NewDerivativeInstrumentNotFoundError creates a derivative-instrument error.
func NewDerivativeInstrumentNotFoundError(msg string) *NSEError {
	return &NSEError{Code: ErrCodeDerivativeInstrumentNotFound, Message: msg}
}
