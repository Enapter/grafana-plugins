package core

import (
	"time"
)

type Timeseries struct {
	TimeField  []time.Time
	DataFields []*TimeseriesDataField
}

func (ts *Timeseries) Len() int {
	return len(ts.TimeField)
}

func (ts *Timeseries) ShiftTime(offset time.Duration) *Timeseries {
	timeField := make([]time.Time, len(ts.TimeField))
	for i, timestamp := range ts.TimeField {
		timeField[i] = timestamp.Add(offset)
	}
	return &Timeseries{
		TimeField:  timeField,
		DataFields: ts.DataFields,
	}
}

type TimeseriesTags map[string]string

type TimeseriesDataField struct {
	Tags   TimeseriesTags
	Type   TimeseriesDataType
	Values []any
}

type TimeseriesDataType uint8

const (
	TimeseriesDataTypeUnknown = iota
	TimeseriesDataTypeFloat
	TimeseriesDataTypeInteger
	TimeseriesDataTypeString
	TimeseriesDataTypeStringArray
	TimeseriesDataTypeBoolean
)

func NewTimeseries(dataTypes []TimeseriesDataType) *Timeseries {
	const preallocValues = 64
	timeField := make([]time.Time, 0, preallocValues)
	dataFields := make([]*TimeseriesDataField, len(dataTypes))
	for i, dataType := range dataTypes {
		dataFields[i] = &TimeseriesDataField{
			Tags:   make(map[string]string),
			Type:   dataType,
			Values: make([]any, 0, preallocValues),
		}
	}
	return &Timeseries{
		TimeField:  timeField,
		DataFields: dataFields,
	}
}
