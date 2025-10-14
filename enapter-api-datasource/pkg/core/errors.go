package core

import (
	"errors"
	"fmt"
	"strings"
)

var ErrTimeseriesEmpty = errors.New("timeseries empty")

type EnapterAPIError struct {
	Code    string
	Message string
	Details map[string]any
}

func (e EnapterAPIError) Error() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("code=%s", e.Code))
	if len(e.Message) > 0 {
		b.WriteString(fmt.Sprintf(", message=%q", e.Message))
	}
	if len(e.Details) > 0 {
		b.WriteString(fmt.Sprintf(", details=%v", e.Details))
	}
	return b.String()
}

var (
	errUnsupportedTimeseriesDataType = errors.New("unsupported timeseries data type")
	errUnexpectedQueryType           = errors.New("unexpected query type")
)

//nolint:stylecheck,revive // user-facing
var (
	ErrSomethingWentWrong = errors.New(
		"Something went wrong. Try again later or contact Enapter support.")
	ErrMetricDataTypeIsNotSupported = errors.New(
		"The requested metric data type is currently not supported.")
	ErrInvalidYAML = errors.New(
		"The query is not a valid YAML.")
	ErrInvalidOffset = errors.New(
		"The offset specified in the query is invalid.")
)
