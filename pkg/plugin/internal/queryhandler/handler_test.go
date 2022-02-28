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
	s.expectGetAndReturnTimeseries(q, &telemetryapi.Timeseries{
		Values:   nil,
		DataType: telemetryapi.TimeseriesDataTypeInt64,
	})
	frames, err := s.handleQuery(q)
	var terr *yaml.TypeError
	s.Require().ErrorAs(err, &terr)
	s.Require().Nil(frames)
}

func (s *QueryHandlerSuite) TestFloat64() {
	q := s.randomDataQuery()
	timeseries := &telemetryapi.Timeseries{
		Values: []*telemetryapi.TimeseriesValue{
			{Timestamp: time.Unix(1, 0), Value: 42.2},
			{Timestamp: time.Unix(2, 0), Value: 43.3},
		},
		DataType: telemetryapi.TimeseriesDataTypeFloat64,
	}
	s.expectGetAndReturnTimeseries(q, timeseries)
	frames, err := s.handleQuery(q)
	s.Require().Nil(err)
	timestampField, valueField := s.extractTimeseriesFields(frames)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal(float64(42.2), valueField.At(0).(float64))
	s.Require().Equal(float64(43.3), valueField.At(1).(float64))
}

func (s *QueryHandlerSuite) TestInt64() {
	q := s.randomDataQuery()
	timeseries := &telemetryapi.Timeseries{
		Values: []*telemetryapi.TimeseriesValue{
			{Timestamp: time.Unix(1, 0), Value: int64(42)},
			{Timestamp: time.Unix(2, 0), Value: int64(43)},
		},
		DataType: telemetryapi.TimeseriesDataTypeInt64,
	}
	s.expectGetAndReturnTimeseries(q, timeseries)
	frames, err := s.handleQuery(q)
	s.Require().Nil(err)
	timestampField, valueField := s.extractTimeseriesFields(frames)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal(int64(42), valueField.At(0).(int64))
	s.Require().Equal(int64(43), valueField.At(1).(int64))
}

func (s *QueryHandlerSuite) TestString() {
	q := s.randomDataQuery()
	timeseries := &telemetryapi.Timeseries{
		Values: []*telemetryapi.TimeseriesValue{
			{Timestamp: time.Unix(1, 0), Value: "foo"},
			{Timestamp: time.Unix(2, 0), Value: "bar"},
		},
		DataType: telemetryapi.TimeseriesDataTypeString,
	}
	s.expectGetAndReturnTimeseries(q, timeseries)
	frames, err := s.handleQuery(q)
	s.Require().Nil(err)
	timestampField, valueField := s.extractTimeseriesFields(frames)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal("foo", valueField.At(0).(string))
	s.Require().Equal("bar", valueField.At(1).(string))
}

func (s *QueryHandlerSuite) TestBool() {
	q := s.randomDataQuery()
	timeseries := &telemetryapi.Timeseries{
		Values: []*telemetryapi.TimeseriesValue{
			{Timestamp: time.Unix(1, 0), Value: true},
			{Timestamp: time.Unix(2, 0), Value: false},
		},
		DataType: telemetryapi.TimeseriesDataTypeBool,
	}
	s.expectGetAndReturnTimeseries(q, timeseries)
	frames, err := s.handleQuery(q)
	s.Require().Nil(err)
	timestampField, valueField := s.extractTimeseriesFields(frames)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal(true, valueField.At(0).(bool))
	s.Require().Equal(false, valueField.At(1).(bool))
}

func (s *QueryHandlerSuite) TestDoNotRenderIntervals() {
	q := s.randomDataQuery()
	q.text = "A fact: $__interval is $__interval_ms milliseconds."
	q.interval = 42 * time.Second
	timeseries := &telemetryapi.Timeseries{
		Values:   nil,
		DataType: telemetryapi.TimeseriesDataTypeBool,
	}
	getParams := q.toGetParams()
	getParams.Query = `{"A fact":"$__interval is $__interval_ms milliseconds."}`
	s.mockTelemetryAPIClient.ExpectGetAndReturn(getParams, timeseries, nil)
	_, err := s.handleQuery(q)
	s.Require().NoError(err)
}

func (s *QueryHandlerSuite) extractTimeseriesFields(frames data.Frames) (
	timestamp *data.Field, value *data.Field,
) {
	s.Require().Len(frames, 1)
	fields := frames[0].Fields
	s.Require().Len(fields, 2)
	timestamp, value = fields[0], fields[1]
	s.Require().Equal("timestamp", timestamp.Name)
	s.Require().Equal("value", value.Name)
	return timestamp, value
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
