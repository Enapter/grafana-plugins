package queryhandler

import "errors"

var (
	errEmptyQueryText                = errors.New("empty query text")
	errUnsupportedTimeseriesDataType = errors.New("unsupported timeseries data type")
	ErrMissingUserInfo               = errors.New("missing user info")
)
