package telemetryapi

import (
	"fmt"
	"strconv"
	"time"
)

type Timeseries struct {
	Values   []*TimeseriesValue
	DataType TimeseriesDataType
}

type TimeseriesValue struct {
	Timestamp time.Time
	Value     interface{}
}

type TimeseriesDataType uint8

const (
	TimeseriesDataTypeUnknown = iota
	TimeseriesDataTypeFloat64
	TimeseriesDataTypeInt64
	TimeseriesDataTypeString
	TimeseriesDataTypeBool
)

func parseTimeseriesDataType(s string) (TimeseriesDataType, error) {
	switch s {
	case "float64":
		return TimeseriesDataTypeFloat64, nil
	case "int64":
		return TimeseriesDataTypeInt64, nil
	case "string":
		return TimeseriesDataTypeString, nil
	case "bool":
		return TimeseriesDataTypeBool, nil
	default:
		return TimeseriesDataTypeUnknown, fmt.Errorf("%w: %s",
			errUnexpectedTimeseriesDataType, s)
	}
}

func (t TimeseriesDataType) String() string {
	switch t {
	case TimeseriesDataTypeFloat64:
		return "float64"
	case TimeseriesDataTypeInt64:
		return "int64"
	case TimeseriesDataTypeString:
		return "string"
	case TimeseriesDataTypeBool:
		return "bool"
	default:
		return "unknown"
	}
}

func (t TimeseriesDataType) Parse(s string) (interface{}, error) {
	switch t {
	case TimeseriesDataTypeFloat64:
		const bitSize = 64
		return strconv.ParseFloat(s, bitSize)
	case TimeseriesDataTypeInt64:
		const base = 10
		const bitSize = 64
		return strconv.ParseInt(s, base, bitSize)
	case TimeseriesDataTypeString:
		return s, nil
	case TimeseriesDataTypeBool:
		return strconv.ParseBool(s)
	default:
		return nil, errUnexpectedTimeseriesDataType
	}
}
