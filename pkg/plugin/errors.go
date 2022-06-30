package plugin

import "errors"

//nolint: stylecheck,revive // user-facing
var (
	errSomethingWentWrong = errors.New(
		"Something went wrong. Try again later or contact Enapter support.")
	errMetricDataTypeIsNotSupported = errors.New(
		"The requested metric data type is currently not supported.")
)
