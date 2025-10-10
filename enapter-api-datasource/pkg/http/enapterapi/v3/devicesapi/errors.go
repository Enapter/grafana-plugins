package devicesapi

import "errors"

var (
	errUnexpectedStatus         = errors.New("unexpected status")
	errUnexpectedContentType    = errors.New("unexpected content type")
	errUnexpectedAbsenceOfError = errors.New("unexpected absence of error")
	ErrNoValues                 = errors.New("no values")
)
