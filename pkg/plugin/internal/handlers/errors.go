package handlers

import "errors"

var (
	errEmptyQueryText                = errors.New("empty query text")
	errUnsupportedTimeseriesDataType = errors.New("unsupported timeseries data type")
)

//nolint: stylecheck,revive // user-facing
var (
	ErrSomethingWentWrong = errors.New(
		"Something went wrong. Try again later or contact Enapter support.")
	ErrMetricDataTypeIsNotSupported = errors.New(
		"The requested metric data type is currently not supported.")
)
