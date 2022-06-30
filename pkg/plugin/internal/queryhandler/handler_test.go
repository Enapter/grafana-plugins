package queryhandler_test

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
	"github.com/stretchr/testify/suite"
	yaml "gopkg.in/yaml.v3"

	"github.com/Enapter/grafana-plugins/telemetry-datasource/pkg/plugin/internal/queryhandler"
	"github.com/Enapter/grafana-plugins/telemetry-datasource/pkg/telemetryapi"
)

type QueryHandlerSuite struct {
	suite.Suite
	ctx                    context.Context
	mockTelemetryAPIClient *MockTelemetryAPIClient
	queryHandler           *queryhandler.QueryHandler
}

func (s *QueryHandlerSuite) SetupSuite() {
	s.ctx = context.Background()
	s.mockTelemetryAPIClient = NewMockTelemetryAPIClient(s.Suite)
	s.queryHandler = queryhandler.New(s.mockTelemetryAPIClient)
}

func (s *QueryHandlerSuite) TestNoUserInfo() {
	pCtx := backend.PluginContext{}
	dataQuery := backend.DataQuery{}
	frames, err := s.queryHandler.HandleQuery(s.ctx, pCtx, dataQuery)
	s.Require().ErrorIs(err, queryhandler.ErrMissingUserInfo)
	s.Require().Nil(frames)
}

var errFake = errors.New("fake error")

func (s *QueryHandlerSuite) TestTelemetryAPIError() {
	q := s.randomDataQuery()
	s.expectGetAndReturnError(q, errFake)
	frames, err := s.handleQuery(q)
	s.Require().ErrorIs(err, errFake)
	s.Require().Nil(frames)
}

func (s *QueryHandlerSuite) TestHandleNoValuesError() {
	q := s.randomDataQuery()
	s.expectGetAndReturnError(q, telemetryapi.ErrNoValues)
	frames, err := s.handleQuery(q)
	s.Require().NoError(err)
	s.Require().Nil(frames)
}

func (s *QueryHandlerSuite) TestEmptyTextNoError() {
	q := s.randomDataQuery()
	q.text = ""
	frames, err := s.handleQuery(q)
	s.Require().NoError(err)
	s.Require().Nil(frames)
}

func (s *QueryHandlerSuite) TestInvalidYAML() {
	q := s.randomDataQuery()
	q.text = "that's not yaml"
	timeseries := telemetryapi.NewTimeseries([]telemetryapi.TimeseriesDataType{
		telemetryapi.TimeseriesDataTypeInt64,
	})
	s.expectGetAndReturnTimeseries(q, timeseries)
	frames, err := s.handleQuery(q)
	var terr *yaml.TypeError
	s.Require().ErrorAs(err, &terr)
	s.Require().Nil(frames)
}

func newFloat64(v float64) *float64 { return &v }
func newInt64(v int64) *int64       { return &v }
func newBool(v bool) *bool          { return &v }
func newString(v string) *string    { return &v }

func (s *QueryHandlerSuite) TestFloat64() {
	q := s.randomDataQuery()
	timeseries := &telemetryapi.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*telemetryapi.TimeseriesDataField{{
			Type: telemetryapi.TimeseriesDataTypeFloat64,
			Values: []interface{}{
				newFloat64(42.2),
				newFloat64(43.3),
			},
		}},
	}
	s.expectGetAndReturnTimeseries(q, timeseries)
	frames, err := s.handleQuery(q)
	s.Require().Nil(err)
	timestampField, dataFields := s.extractTimeseriesFields(frames)
	s.Require().Len(dataFields, 1)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal(float64(42.2), *dataFields[0].At(0).(*float64))
	s.Require().Equal(float64(43.3), *dataFields[0].At(1).(*float64))
}

func (s *QueryHandlerSuite) TestInt64() {
	q := s.randomDataQuery()
	timeseries := &telemetryapi.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*telemetryapi.TimeseriesDataField{{
			Type: telemetryapi.TimeseriesDataTypeInt64,
			Values: []interface{}{
				newInt64(42),
				newInt64(43),
			},
		}},
	}
	s.expectGetAndReturnTimeseries(q, timeseries)
	frames, err := s.handleQuery(q)
	s.Require().Nil(err)
	timestampField, dataFields := s.extractTimeseriesFields(frames)
	s.Require().Len(dataFields, 1)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal(int64(42), *dataFields[0].At(0).(*int64))
	s.Require().Equal(int64(43), *dataFields[0].At(1).(*int64))
}

func (s *QueryHandlerSuite) TestString() {
	q := s.randomDataQuery()
	timeseries := &telemetryapi.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*telemetryapi.TimeseriesDataField{{
			Type: telemetryapi.TimeseriesDataTypeString,
			Values: []interface{}{
				newString("foo"),
				newString("bar"),
			},
		}},
	}
	s.expectGetAndReturnTimeseries(q, timeseries)
	frames, err := s.handleQuery(q)
	s.Require().Nil(err)
	timestampField, dataFields := s.extractTimeseriesFields(frames)
	s.Require().Len(dataFields, 1)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal("foo", *dataFields[0].At(0).(*string))
	s.Require().Equal("bar", *dataFields[0].At(1).(*string))
}

func (s *QueryHandlerSuite) TestStringArrayIsUnsupported() {
	q := s.randomDataQuery()
	timeseries := &telemetryapi.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*telemetryapi.TimeseriesDataField{{
			Type: telemetryapi.TimeseriesDataTypeStringArray,
			Values: []interface{}{
				[]string{"foo", "foo"},
				[]string{"bar", "bar"},
			},
		}},
	}
	s.expectGetAndReturnTimeseries(q, timeseries)
	_, err := s.handleQuery(q)
	s.Require().ErrorIs(err, queryhandler.ErrUnsupportedTimeseriesDataType)
}

func (s *QueryHandlerSuite) TestBool() {
	q := s.randomDataQuery()
	timeseries := &telemetryapi.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*telemetryapi.TimeseriesDataField{{
			Type: telemetryapi.TimeseriesDataTypeBool,
			Values: []interface{}{
				newBool(true),
				newBool(false),
			},
		}},
	}
	s.expectGetAndReturnTimeseries(q, timeseries)
	frames, err := s.handleQuery(q)
	s.Require().Nil(err)
	timestampField, dataFields := s.extractTimeseriesFields(frames)
	s.Require().Len(dataFields, 1)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal(true, *dataFields[0].At(0).(*bool))
	s.Require().Equal(false, *dataFields[0].At(1).(*bool))
}

func (s *QueryHandlerSuite) TestMultipleFields() {
	q := s.randomDataQuery()
	timeseries := &telemetryapi.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*telemetryapi.TimeseriesDataField{
			{
				Type: telemetryapi.TimeseriesDataTypeFloat64,
				Values: []interface{}{
					newFloat64(42.2),
					newFloat64(43.3),
				},
			},
			{
				Type: telemetryapi.TimeseriesDataTypeString,
				Values: []interface{}{
					newString("foo"),
					newString("bar"),
				},
			},
		},
	}
	s.expectGetAndReturnTimeseries(q, timeseries)
	frames, err := s.handleQuery(q)
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

func (s *QueryHandlerSuite) TestMultipleFieldsWithNil() {
	q := s.randomDataQuery()
	timeseries := &telemetryapi.Timeseries{
		TimeField: []time.Time{
			time.Unix(1, 0),
			time.Unix(2, 0),
		},
		DataFields: []*telemetryapi.TimeseriesDataField{
			{
				Type: telemetryapi.TimeseriesDataTypeFloat64,
				Values: []interface{}{
					newFloat64(42.2),
					(*float64)(nil),
				},
			},
			{
				Type: telemetryapi.TimeseriesDataTypeString,
				Values: []interface{}{
					newString("foo"),
					newString("bar"),
				},
			},
		},
	}
	s.expectGetAndReturnTimeseries(q, timeseries)
	frames, err := s.handleQuery(q)
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

func (s *QueryHandlerSuite) TestDoNotRenderIntervals() {
	q := s.randomDataQuery()
	q.text = "A fact: $__interval is $__interval_ms milliseconds."
	q.interval = 42 * time.Second
	timeseries := telemetryapi.NewTimeseries([]telemetryapi.TimeseriesDataType{
		telemetryapi.TimeseriesDataTypeBool,
	})
	getParams := q.toGetParams()
	getParams.Query = `{"A fact":"$__interval is $__interval_ms milliseconds."}`
	s.mockTelemetryAPIClient.ExpectGetAndReturn(getParams, timeseries, nil)
	_, err := s.handleQuery(q)
	s.Require().NoError(err)
}

func (s *QueryHandlerSuite) extractTimeseriesFields(frames data.Frames) (
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

func (s *QueryHandlerSuite) expectGetAndReturnTimeseries(q dataQuery, ts *telemetryapi.Timeseries) {
	s.mockTelemetryAPIClient.ExpectGetAndReturn(q.toGetParams(), ts, nil)
}

func (s *QueryHandlerSuite) expectGetAndReturnError(q dataQuery, err error) {
	s.mockTelemetryAPIClient.ExpectGetAndReturn(q.toGetParams(), nil, err)
}

func (s *QueryHandlerSuite) randomDataQuery() dataQuery {
	return dataQuery{
		user:     faker.Email(),
		from:     time.Now().Add(-time.Duration(rand.Int()+1) * time.Hour),
		to:       time.Now().Add(-time.Duration(rand.Int()+1) * time.Minute),
		interval: time.Duration(rand.Int()) * time.Second,
		text: string(s.shouldMarshalJSON(map[string]interface{}{
			faker.Word(): faker.Sentence(),
			faker.Word(): rand.Int(),
			faker.Word(): rand.Int()%2 == 0,
		})),
	}
}

func (s *QueryHandlerSuite) handleQuery(q dataQuery) (data.Frames, error) {
	pCtx := backend.PluginContext{
		User: &backend.User{
			Email: q.user,
		},
	}
	dataQuery := backend.DataQuery{
		TimeRange: backend.TimeRange{From: q.from, To: q.to},
		Interval:  q.interval,
		JSON: s.shouldMarshalJSON(map[string]interface{}{
			"text": q.text,
		}),
	}
	return s.queryHandler.HandleQuery(s.ctx, pCtx, dataQuery)
}

func (s *QueryHandlerSuite) shouldMarshalJSON(o interface{}) []byte {
	data, err := json.Marshal(o)
	s.Require().NoError(err)
	return data
}

type dataQuery struct {
	user     string
	from     time.Time
	to       time.Time
	interval time.Duration
	text     string
}

func (q dataQuery) toGetParams() telemetryapi.TimeseriesParams {
	return telemetryapi.TimeseriesParams{
		User:  q.user,
		Query: q.text,
		From:  q.from,
		To:    q.to,
	}
}

func TestQueryHandler(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(QueryHandlerSuite))
}
