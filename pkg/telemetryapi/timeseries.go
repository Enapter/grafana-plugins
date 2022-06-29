package telemetryapi

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Timeseries struct {
	TimeField  []time.Time
	DataFields []*TimeseriesDataField
}

func (ts *Timeseries) Len() int {
	return len(ts.TimeField)
}

type TimeseriesTags map[string]string

func parseTimeseriesTags(s string) (TimeseriesTags, error) {
	tags := make(TimeseriesTags)

	pairs := strings.Split(s, " ")

	for _, pair := range pairs {
		kv := strings.Split(pair, "=")

		if want := 2; len(kv) != want {
			return nil, fmt.Errorf("%w: len: want %d, have %d",
				errBadKeyValuePair, want, len(kv))
		}

		tags[kv[0]] = kv[1]
	}

	return tags, nil
}

type TimeseriesDataField struct {
	Tags   TimeseriesTags
	Type   TimeseriesDataType
	Values []interface{}
}

type TimeseriesDataType uint8

const (
	TimeseriesDataTypeUnknown = iota
	TimeseriesDataTypeFloat64
	TimeseriesDataTypeInt64
	TimeseriesDataTypeString
	TimeseriesDataTypeBool
)

func parseTimeseriesDataTypes(ss []string) ([]TimeseriesDataType, error) {
	dataTypes := make([]TimeseriesDataType, len(ss))

	for i, s := range ss {
		t, err := parseTimeseriesDataType(s)
		if err != nil {
			return nil, fmt.Errorf("%d: %w", i, err)
		}
		dataTypes[i] = t
	}

	return dataTypes, nil
}

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

func NewTimeseries(dataTypes []TimeseriesDataType) *Timeseries {
	const preallocValues = 64

	timeField := make([]time.Time, 0, preallocValues)

	dataFields := make([]*TimeseriesDataField, len(dataTypes))
	for i, dataType := range dataTypes {
		dataFields[i] = &TimeseriesDataField{
			Tags:   make(map[string]string),
			Type:   dataType,
			Values: make([]interface{}, 0, preallocValues),
		}
	}

	return &Timeseries{
		TimeField:  timeField,
		DataFields: dataFields,
	}
}
