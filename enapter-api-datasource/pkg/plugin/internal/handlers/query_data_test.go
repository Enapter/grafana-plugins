package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"

	"github.com/Enapter/grafana-plugins/pkg/core"
	"github.com/Enapter/grafana-plugins/pkg/plugin/internal/handlers"
)

type QueryDataSuite struct {
	suite.Suite
	ctx                   context.Context
	logger                hclog.Logger
	mockEnapterAPIAdapter *MockEnapterAPIAdapter
	queryDataHandler      *handlers.QueryData
}

func (s *QueryDataSuite) SetupSuite() {
	s.ctx = context.Background()
	s.mockEnapterAPIAdapter = NewMockEnapterAPIAdapter(&s.Suite)
	s.logger = hclog.Default()
	s.queryDataHandler = handlers.NewQueryData(s.logger, s.mockEnapterAPIAdapter)
}

var errFake = errors.New("fake error")

func (s *QueryDataSuite) TestDefaultGranularity() {
	for in, out := range map[time.Duration]time.Duration{
		999 * time.Millisecond:  time.Second,
		1000 * time.Millisecond: time.Second,
		1001 * time.Millisecond: 2 * time.Second,
		1999 * time.Millisecond: 2 * time.Second,
		2000 * time.Millisecond: 2 * time.Second,
		2001 * time.Millisecond: 5 * time.Second,
		4 * time.Second:         5 * time.Second,
		5 * time.Second:         5 * time.Second,
		6 * time.Second:         time.Minute,
		59 * time.Second:        time.Minute,
		60 * time.Second:        time.Minute,
		61 * time.Second:        2 * time.Minute,
		119 * time.Second:       2 * time.Minute,
		120 * time.Second:       2 * time.Minute,
		121 * time.Second:       5 * time.Minute,
		4 * time.Minute:         5 * time.Minute,
		5 * time.Minute:         5 * time.Minute,
		6 * time.Minute:         10 * time.Minute,
		9 * time.Minute:         10 * time.Minute,
		10 * time.Minute:        10 * time.Minute,
		11 * time.Minute:        20 * time.Minute,
		19 * time.Minute:        20 * time.Minute,
		20 * time.Minute:        20 * time.Minute,
		21 * time.Minute:        30 * time.Minute,
		29 * time.Minute:        30 * time.Minute,
		30 * time.Minute:        30 * time.Minute,
		31 * time.Minute:        time.Hour,
		59 * time.Minute:        time.Hour,
		60 * time.Minute:        time.Hour,
		61 * time.Minute:        2 * time.Hour,
		119 * time.Minute:       2 * time.Hour,
		120 * time.Minute:       2 * time.Hour,
		121 * time.Minute:       6 * time.Hour,
		5 * time.Hour:           6 * time.Hour,
		6 * time.Hour:           6 * time.Hour,
		7 * time.Hour:           12 * time.Hour,
		11 * time.Hour:          12 * time.Hour,
		12 * time.Hour:          12 * time.Hour,
		13 * time.Hour:          24 * time.Hour,
		23 * time.Hour:          24 * time.Hour,
		24 * time.Hour:          24 * time.Hour,
		25 * time.Hour:          24 * time.Hour,
	} {
		in := in
		out := out
		s.Run(in.String(), func() {
			gran := s.queryDataHandler.DefaultGranularity(in)
			s.Require().Equal(out, gran)
		})
	}
}

func (s *QueryDataSuite) TestCommandRequest() {
	req := s.randomDataRequestWithSingleCommandQuery()
	stateIn := faker.Word()
	payloadIn := map[string]interface{}{
		faker.Word(): faker.Word(),
		faker.Word(): faker.Word(),
	}
	s.expectExecuteCommandAndReturn(req, &core.ExecuteCommandResponse{
		State:   stateIn,
		Payload: payloadIn,
	}, nil)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().NoError(err)
	stateOut, payloadOut := s.extractCommandResponse(frames)
	s.Require().Equal(stateIn, stateOut)
	s.Require().Equal(payloadIn, payloadOut)
}

func (s *QueryDataSuite) TestTelemetryAPIError() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	s.expectQueryTimeseriesAndReturn(req, nil, errFake)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().ErrorIs(err, handlers.ErrSomethingWentWrong)
	s.Require().Nil(frames)
}

func (s *QueryDataSuite) TestHandleNoValuesError() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	s.expectQueryTimeseriesAndReturn(req, nil, core.ErrTimeseriesEmpty)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().NoError(err)
	s.Require().Nil(frames)
}

func (s *QueryDataSuite) TestEmptyTextNoError() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	req.queries[0].text = ""
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().NoError(err)
	s.Require().Nil(frames)
}

func (s *QueryDataSuite) TestInvalidYAML() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	req.queries[0].text = "that's not yaml"
	timeseries := core.NewTimeseries([]core.TimeseriesDataType{
		core.TimeseriesDataTypeInteger,
	})
	s.expectQueryTimeseriesAndReturn(req, &core.QueryTimeseriesResponse{
		Timeseries: timeseries,
	}, nil)
	_, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().ErrorIs(err, handlers.ErrInvalidYAML)
}

func (s *QueryDataSuite) TestEnapterAPIError() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	timeseries := core.NewTimeseries([]core.TimeseriesDataType{
		core.TimeseriesDataTypeInteger,
	})
	s.expectQueryTimeseriesAndReturn(req, &core.QueryTimeseriesResponse{
		Timeseries: timeseries,
	}, core.EnapterAPIError{
		Code:    "apollo_13",
		Message: "Houston, we have a problem.",
		Details: map[string]any{
			"command_module_pilot": "Jack Swigert",
		},
	})
	_, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Error(err)
	s.Require().Equal("Houston, we have a problem.", err.Error())
}

func (s *QueryDataSuite) TestInvalidOffset() {
	req := dataRequest{
		user: faker.Email(),
		queries: []query{{
			refID:    s.randomRefID(),
			from:     time.Now().Add(-time.Duration(rand.Int()+1) * time.Hour),
			to:       time.Now().Add(-time.Duration(rand.Int()+1) * time.Minute),
			interval: time.Duration(rand.Int()) * time.Second,
			text: string(s.shouldMarshalJSON(map[string]interface{}{
				"@offset": "42",
			})),
		}},
	}
	_, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().ErrorIs(err, handlers.ErrInvalidOffset)
}

func newFloat64(v float64) *float64 { return &v }
func newInt64(v int64) *int64       { return &v }
func newBool(v bool) *bool          { return &v }
func newString(v string) *string    { return &v }

//nolint:dupl // FIXME
func (s *QueryDataSuite) TestFloat64() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	timeseries := &core.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*core.TimeseriesDataField{{
			Type: core.TimeseriesDataTypeFloat,
			Values: []interface{}{
				newFloat64(42.2),
				newFloat64(43.3),
			},
		}},
	}
	s.expectQueryTimeseriesAndReturn(req, &core.QueryTimeseriesResponse{
		Timeseries: timeseries,
	}, nil)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Nil(err)
	timestampField, dataFields := s.extractTimeseriesFields(frames)
	s.Require().Len(dataFields, 1)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal(float64(42.2), *dataFields[0].At(0).(*float64))
	s.Require().Equal(float64(43.3), *dataFields[0].At(1).(*float64))
}

//nolint:dupl // FIXME
func (s *QueryDataSuite) TestInt64() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	timeseries := &core.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*core.TimeseriesDataField{{
			Type: core.TimeseriesDataTypeInteger,
			Values: []interface{}{
				newInt64(42),
				newInt64(43),
			},
		}},
	}
	s.expectQueryTimeseriesAndReturn(req, &core.QueryTimeseriesResponse{
		Timeseries: timeseries,
	}, nil)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Nil(err)
	timestampField, dataFields := s.extractTimeseriesFields(frames)
	s.Require().Len(dataFields, 1)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal(int64(42), *dataFields[0].At(0).(*int64))
	s.Require().Equal(int64(43), *dataFields[0].At(1).(*int64))
}

func (s *QueryDataSuite) TestHide() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	req.queries[0].hide = true
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Nil(err)
	s.Require().Nil(frames)
}

func (s *QueryDataSuite) TestString() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	timeseries := &core.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*core.TimeseriesDataField{{
			Type: core.TimeseriesDataTypeString,
			Values: []interface{}{
				newString("foo"),
				newString("bar"),
			},
		}},
	}
	s.expectQueryTimeseriesAndReturn(req, &core.QueryTimeseriesResponse{
		Timeseries: timeseries,
	}, nil)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Nil(err)
	timestampField, dataFields := s.extractTimeseriesFields(frames)
	s.Require().Len(dataFields, 1)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal("foo", *dataFields[0].At(0).(*string))
	s.Require().Equal("bar", *dataFields[0].At(1).(*string))
}

func (s *QueryDataSuite) TestOffset() {
	req := dataRequest{
		user: faker.Email(),
		queries: []query{{
			refID:    s.randomRefID(),
			from:     time.Unix(5, 0),
			to:       time.Unix(10, 0),
			interval: time.Second,
			text: string(s.shouldMarshalJSON(map[string]interface{}{
				"granularity": "42s",
				"aggregation": "auto",
				"@offset":     "2s",
			})),
		}},
	}
	timeseries := &core.Timeseries{
		TimeField: []time.Time{
			time.Unix(3, 0),
			time.Unix(4, 0),
			time.Unix(5, 0),
			time.Unix(6, 0),
			time.Unix(7, 0),
		},
		DataFields: []*core.TimeseriesDataField{{
			Type: core.TimeseriesDataTypeInteger,
			Values: []interface{}{
				newInt64(1),
				newInt64(2),
				newInt64(3),
				newInt64(4),
				newInt64(5),
			},
		}},
	}
	s.mockEnapterAPIAdapter.ExpectQueryTimeseriesAndReturn(
		&core.QueryTimeseriesRequest{
			User: req.user,
			Query: string(s.shouldMarshalJSON(map[string]interface{}{
				"aggregation": "auto",
				"from":        "1970-01-01T00:00:03Z",
				"granularity": "42s",
				"to":          "1970-01-01T00:00:08Z",
			})),
		}, &core.QueryTimeseriesResponse{Timeseries: timeseries}, nil)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Nil(err)
	timestampField, dataFields := s.extractTimeseriesFields(frames)
	s.Require().Len(dataFields, 1)
	s.Require().Equal(int64(5), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(6), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal(int64(7), timestampField.At(2).(time.Time).Unix())
	s.Require().Equal(int64(8), timestampField.At(3).(time.Time).Unix())
	s.Require().Equal(int64(9), timestampField.At(4).(time.Time).Unix())
	s.Require().Equal(int64(1), *dataFields[0].At(0).(*int64))
	s.Require().Equal(int64(2), *dataFields[0].At(1).(*int64))
	s.Require().Equal(int64(3), *dataFields[0].At(2).(*int64))
	s.Require().Equal(int64(4), *dataFields[0].At(3).(*int64))
	s.Require().Equal(int64(5), *dataFields[0].At(4).(*int64))
}

func (s *QueryDataSuite) TestNoUserInfo() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	req.user = ""
	timeseries := &core.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*core.TimeseriesDataField{{
			Type: core.TimeseriesDataTypeString,
			Values: []interface{}{
				newString("foo"),
				newString("bar"),
			},
		}},
	}
	s.expectQueryTimeseriesAndReturn(req, &core.QueryTimeseriesResponse{
		Timeseries: timeseries,
	}, nil)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Nil(err)
	timestampField, dataFields := s.extractTimeseriesFields(frames)
	s.Require().Len(dataFields, 1)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal("foo", *dataFields[0].At(0).(*string))
	s.Require().Equal("bar", *dataFields[0].At(1).(*string))
}

func (s *QueryDataSuite) TestStringArrayIsUnsupported() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	timeseries := &core.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*core.TimeseriesDataField{{
			Type: core.TimeseriesDataTypeStringArray,
			Values: []interface{}{
				[]string{"foo", "foo"},
				[]string{"bar", "bar"},
			},
		}},
	}
	s.expectQueryTimeseriesAndReturn(req, &core.QueryTimeseriesResponse{
		Timeseries: timeseries,
	}, nil)
	_, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Error(err, handlers.ErrMetricDataTypeIsNotSupported)
}

func (s *QueryDataSuite) TestBool() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	timeseries := &core.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*core.TimeseriesDataField{{
			Type: core.TimeseriesDataTypeBoolean,
			Values: []interface{}{
				newBool(true),
				newBool(false),
			},
		}},
	}
	s.expectQueryTimeseriesAndReturn(req, &core.QueryTimeseriesResponse{
		Timeseries: timeseries,
	}, nil)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Nil(err)
	timestampField, dataFields := s.extractTimeseriesFields(frames)
	s.Require().Len(dataFields, 1)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal(true, *dataFields[0].At(0).(*bool))
	s.Require().Equal(false, *dataFields[0].At(1).(*bool))
}

func (s *QueryDataSuite) TestMultipleFields() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	timeseries := &core.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*core.TimeseriesDataField{
			{
				Type: core.TimeseriesDataTypeFloat,
				Values: []interface{}{
					newFloat64(42.2),
					newFloat64(43.3),
				},
			},
			{
				Type: core.TimeseriesDataTypeString,
				Values: []interface{}{
					newString("foo"),
					newString("bar"),
				},
			},
		},
	}
	s.expectQueryTimeseriesAndReturn(req, &core.QueryTimeseriesResponse{
		Timeseries: timeseries,
	}, nil)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Nil(err)
	timestampField, dataFields := s.extractTimeseriesFields(frames)
	s.Require().Len(dataFields, 2)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal(float64(42.2), *dataFields[0].At(0).(*float64))
	s.Require().Equal(float64(43.3), *dataFields[0].At(1).(*float64))
	s.Require().Equal("foo", *dataFields[1].At(0).(*string))
	s.Require().Equal("bar", *dataFields[1].At(1).(*string))
}

func (s *QueryDataSuite) TestMultipleFieldsWithNil() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	timeseries := &core.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*core.TimeseriesDataField{
			{
				Type: core.TimeseriesDataTypeFloat,
				Values: []interface{}{
					newFloat64(42.2),
					(*float64)(nil),
				},
			},
			{
				Type: core.TimeseriesDataTypeString,
				Values: []interface{}{
					newString("foo"),
					newString("bar"),
				},
			},
		},
	}
	s.expectQueryTimeseriesAndReturn(req, &core.QueryTimeseriesResponse{
		Timeseries: timeseries,
	}, nil)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Nil(err)
	timestampField, dataFields := s.extractTimeseriesFields(frames)
	s.Require().Len(dataFields, 2)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal(float64(42.2), *dataFields[0].At(0).(*float64))
	s.Require().Equal((*float64)(nil), dataFields[0].At(1).(*float64))
	s.Require().Equal("foo", *dataFields[1].At(0).(*string))
	s.Require().Equal("bar", *dataFields[1].At(1).(*string))
}

func (s *QueryDataSuite) TestDoNotRenderIntervals() {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	req.queries[0].text = `{"A fact":"$__interval is $__interval_ms milliseconds.",` +
		`"granularity":"42s","aggregation":"avg"}`
	req.queries[0].interval = time.Duration(rand.Int()+1) * time.Second
	timeseries := core.NewTimeseries([]core.TimeseriesDataType{
		core.TimeseriesDataTypeBoolean,
	})
	s.expectQueryTimeseriesAndReturn(req, &core.QueryTimeseriesResponse{
		Timeseries: timeseries,
	}, nil)
	_, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().NoError(err)
}

func (s *QueryDataSuite) TestUniqueLabelsEmpty() {
	s.testUniqueLabels([]core.TimeseriesTags{
		{"foo": "bar"},
	}, []data.Labels{
		{},
	})
}

func (s *QueryDataSuite) TestUniqueLabelsNoop() {
	s.testUniqueLabels([]core.TimeseriesTags{
		{"foo": "bar"},
		{"goo": "jar"},
	}, []data.Labels{
		{"foo": "bar"},
		{"goo": "jar"},
	})
}

func (s *QueryDataSuite) TestUniqueLabelsRemoveDuplicate() {
	s.testUniqueLabels([]core.TimeseriesTags{
		{"foo": "bar", "goo": "jar"},
		{"foo": "bar", "goo": "JAR"},
	}, []data.Labels{
		{"goo": "jar"},
		{"goo": "JAR"},
	})
}

func (s *QueryDataSuite) TestUniqueLabelsDefaultName() {
	s.testUniqueLabels([]core.TimeseriesTags{
		{"foo": "bar", "telemetry": "h2_flow"},
	}, []data.Labels{
		{"telemetry": "h2_flow"},
	})
}

func (s *QueryDataSuite) testUniqueLabels(
	tagsIn []core.TimeseriesTags, labelsOut []data.Labels,
) {
	req := s.randomDataRequestWithSingleTelemetryQuery()
	dataFieldsIn := make([]*core.TimeseriesDataField, len(tagsIn))
	for i, tags := range tagsIn {
		dataFieldsIn[i] = &core.TimeseriesDataField{
			Tags: tags,
			Type: core.TimeseriesDataTypeFloat,
		}
	}
	timeseries := &core.Timeseries{
		TimeField:  []time.Time{},
		DataFields: dataFieldsIn,
	}
	s.expectQueryTimeseriesAndReturn(req, &core.QueryTimeseriesResponse{
		Timeseries: timeseries,
	}, nil)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Nil(err)
	_, dataFieldsOut := s.extractTimeseriesFields(frames)
	for i, field := range dataFieldsOut {
		s.Require().Equal(labelsOut[i], field.Labels)
	}
}

func (s *QueryDataSuite) extractTimeseriesFields(frames data.Frames) (
	timestamp *data.Field, dataFields []*data.Field,
) {
	s.Require().Len(frames, 1)
	fields := frames[0].Fields
	time, values := fields[0], fields[1:]
	s.Require().Equal("time", time.Name)
	for _, value := range values {
		s.Require().Equal("", value.Name)
	}
	return time, values
}

func (s *QueryDataSuite) extractCommandResponse(
	frames data.Frames,
) (state string, payload map[string]interface{}) {
	s.Require().Len(frames, 1)
	fields := frames[0].Fields

	const nFields = 2
	s.Require().Len(fields, nFields)

	stateField := fields[0]
	s.Require().Equal(1, stateField.Len())
	state = stateField.At(0).(string)

	payloadField := fields[1]
	s.Require().Equal(1, payloadField.Len())
	err := json.Unmarshal(payloadField.At(0).(json.RawMessage), &payload)
	s.Require().NoError(err)

	return state, payload
}

func (s *QueryDataSuite) expectQueryTimeseriesAndReturn(
	req dataRequest, resp *core.QueryTimeseriesResponse, err error,
) {
	for _, q := range req.queries {
		s.mockEnapterAPIAdapter.ExpectQueryTimeseriesAndReturn(
			&core.QueryTimeseriesRequest{
				User:  req.user,
				Query: s.queryTextWithTimeRange(q),
			}, resp, err)
	}
}

func (s *QueryDataSuite) expectExecuteCommandAndReturn(
	req dataRequest, resp *core.ExecuteCommandResponse, err error,
) {
	for _, q := range req.queries {
		s.mockEnapterAPIAdapter.ExpectExecuteCommandAndReturn(
			&core.ExecuteCommandRequest{
				User:        req.user,
				CommandName: q.payload["commandName"].(string),
				CommandArgs: q.payload["commandArgs"].(map[string]interface{}),
				DeviceID:    q.payload["deviceId"].(string),
			}, resp, err)
	}
}

func (s *QueryDataSuite) queryTextWithTimeRange(q query) string {
	var obj map[string]interface{}
	if err := yaml.Unmarshal([]byte(q.text), &obj); err == nil {
		obj["from"] = q.from.UTC().Format(time.RFC3339Nano)
		obj["to"] = q.to.UTC().Format(time.RFC3339Nano)
		delete(obj, "@offset")
	}

	out, err := json.Marshal(obj)
	s.Require().NoError(err)
	return string(out)
}

func (s *QueryDataSuite) randomDataRequestWithSingleTelemetryQuery() dataRequest {
	return dataRequest{
		user: faker.Email(),
		queries: []query{{
			refID:    s.randomRefID(),
			from:     time.Now().Add(-time.Duration(rand.Int()+1) * time.Hour),
			to:       time.Now().Add(-time.Duration(rand.Int()+1) * time.Minute),
			interval: time.Duration(rand.Int()) * time.Second,
			text: string(s.shouldMarshalJSON(map[string]interface{}{
				faker.Word():  faker.Sentence(),
				faker.Word():  rand.Int(),
				faker.Word():  rand.Int()%2 == 0,
				"granularity": "42s",
				"aggregation": "auto",
			})),
		}},
	}
}

func (s *QueryDataSuite) randomDataRequestWithSingleCommandQuery() dataRequest {
	return dataRequest{
		user: faker.Email(),
		queries: []query{{
			refID:     s.randomRefID(),
			queryType: "command",
			payload: map[string]interface{}{
				"commandName": faker.Word(),
				"commandArgs": map[string]interface{}{
					faker.Word(): faker.Word(),
					faker.Word(): faker.Word(),
				},
				"deviceId": faker.Word(),
			},
		}},
	}
}

func (s *QueryDataSuite) randomRefID() string {
	const abc = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return string(abc[rand.Int()%len(abc)])
}

func (s *QueryDataSuite) handleDataRequestWithSingleQuery(
	req dataRequest,
) (data.Frames, error) {
	s.Require().Len(req.queries, 1)

	responses := s.handleDataRequest(req)
	s.Require().Len(responses, len(req.queries))

	resp := responses[req.queries[0].refID]
	return resp.Frames, resp.Error
}

func (s *QueryDataSuite) handleDataRequest(req dataRequest) backend.Responses {
	var user *backend.User
	if req.user != "" {
		user = &backend.User{
			Email: req.user,
		}
	}

	queries := make([]backend.DataQuery, len(req.queries))
	for i, q := range req.queries {
		queries[i] = backend.DataQuery{
			RefID:     q.refID,
			QueryType: q.queryType,
			TimeRange: backend.TimeRange{From: q.from.UTC(), To: q.to.UTC()},
			Interval:  q.interval,
			JSON: s.shouldMarshalJSON(map[string]interface{}{
				"text":    q.text,
				"hide":    q.hide,
				"payload": q.payload,
			}),
		}
	}

	resp, err := s.queryDataHandler.QueryData(s.ctx, &backend.QueryDataRequest{
		PluginContext: backend.PluginContext{
			User: user,
		},
		Headers: nil,
		Queries: queries,
	})
	s.Require().NoError(err)

	return resp.Responses
}

func (s *QueryDataSuite) shouldMarshalJSON(o interface{}) []byte {
	data, err := json.Marshal(o)
	s.Require().NoError(err)
	return data
}

type dataRequest struct {
	user    string
	queries []query
}

type query struct {
	refID     string
	queryType string
	from      time.Time
	to        time.Time
	interval  time.Duration
	hide      bool
	text      string
	payload   map[string]interface{}
}

func TestQueryData(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(QueryDataSuite))
}
