package telemetryapi

import "errors"

var (
	ErrEmptyUser                    = errors.New("empty user")
	errEmptyBaseURL                 = errors.New("empty base URL")
	errEmptyHeaderField             = errors.New("empty header field")
	errUnexpectedStatus             = errors.New("unexpected status")
	errUnexpectedNumberOfFields     = errors.New("unexpected number of fields")
	errUnexpectedContentType        = errors.New("unexpected content type")
	errUnexpectedTimeseriesDataType = errors.New("unexpected timeseries data type")
	errUnexpectedCSVHeader          = errors.New("unexpected csv header")
	errEmptyData                    = errors.New("empty data")
	errEmptyErrorList               = errors.New("empty error list")
	ErrNoValues                     = errors.New("no values")
)
