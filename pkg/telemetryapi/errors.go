package telemetryapi

import "errors"

var (
	errEmptyHeaderField             = errors.New("empty header field")
	errUnexpectedStatus             = errors.New("unexpected status")
	errUnexpectedAbsenceOfError     = errors.New("unexpected absence of error")
	errUnexpectedNumberOfFields     = errors.New("unexpected number of fields")
	errUnexpectedContentType        = errors.New("unexpected content type")
	errUnexpectedTimeseriesDataType = errors.New("unexpected timeseries data type")
	errUnexpectedFieldName          = errors.New("unexpected field name")
	errEmptyData                    = errors.New("empty data")
	errEmptyErrorList               = errors.New("empty error list")
	ErrNoValues                     = errors.New("no values")
	errBadKeyValuePair              = errors.New("bad key value pair")
)
