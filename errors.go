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

func (e *NSEError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Convenience constructors for each error type.
func NewAPIError(msg string) *NSEError { return &NSEError{Code: ErrCodeAPI, Message: msg} }
func NewDataNotFoundError(msg string) *NSEError {
	return &NSEError{Code: ErrCodeDataNotFound, Message: msg}
}
func NewIndexDataNotFoundError(msg string) *NSEError {
	return &NSEError{Code: ErrCodeIndexDataNotFound, Message: msg}
}
func NewCalendarNotFoundError(msg string) *NSEError {
	return &NSEError{Code: ErrCodeCalendarNotFound, Message: msg}
}
func NewInvalidIndexCategoryError(msg string) *NSEError {
	return &NSEError{Code: ErrCodeInvalidIndexCategory, Message: msg}
}
func NewInvalidIndexError(msg string) *NSEError {
	return &NSEError{Code: ErrCodeInvalidIndex, Message: msg}
}
func NewDerivativeInstrumentNotFoundError(msg string) *NSEError {
	return &NSEError{Code: ErrCodeDerivativeInstrumentNotFound, Message: msg}
}
