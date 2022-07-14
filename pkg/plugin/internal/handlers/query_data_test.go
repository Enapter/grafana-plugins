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
	yaml "gopkg.in/yaml.v3"

	"github.com/Enapter/grafana-plugins/telemetry-datasource/pkg/plugin/internal/handlers"
	"github.com/Enapter/grafana-plugins/telemetry-datasource/pkg/telemetryapi"
)

type QueryDataSuite struct {
	suite.Suite
	ctx                    context.Context
	logger                 hclog.Logger
	mockTelemetryAPIClient *MockTelemetryAPIClient
	queryHandler           *handlers.QueryData
}

func (s *QueryDataSuite) SetupSuite() {
	s.ctx = context.Background()
	s.mockTelemetryAPIClient = NewMockTelemetryAPIClient(s.Suite)
	s.logger = hclog.Default()
	s.queryHandler = handlers.NewQueryData(s.logger, s.mockTelemetryAPIClient)
}

var errFake = errors.New("fake error")

func (s *QueryDataSuite) TestTelemetryAPIError() {
	q := s.randomDataQuery()
	s.expectGetAndReturnError(q, errFake)
	frames, err := s.handleQuery(q)
	s.Require().ErrorIs(err, errFake)
	s.Require().Nil(frames)
}

func (s *QueryDataSuite) TestHandleNoValuesError() {
	q := s.randomDataQuery()
	s.expectGetAndReturnError(q, telemetryapi.ErrNoValues)
	frames, err := s.handleQuery(q)
	s.Require().NoError(err)
	s.Require().Nil(frames)
}

func (s *QueryDataSuite) TestEmptyTextNoError() {
	q := s.randomDataQuery()
	q.text = ""
	frames, err := s.handleQuery(q)
	s.Require().NoError(err)
	s.Require().Nil(frames)
}

func (s *QueryDataSuite) TestInvalidYAML() {
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

//nolint: dupl // FIXME
func (s *QueryDataSuite) TestFloat64() {
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

//nolint: dupl // FIXME
func (s *QueryDataSuite) TestInt64() {
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

func (s *QueryDataSuite) TestHide() {
	q := s.randomDataQuery()
	q.hide = true
	frames, err := s.handleQuery(q)
	s.Require().Nil(err)
	s.Require().Nil(frames)
}

func (s *QueryDataSuite) TestString() {
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

func (s *QueryDataSuite) TestNoUserInfo() {
	q := s.randomDataQuery()
	q.user = ""
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

func (s *QueryDataSuite) TestStringArrayIsUnsupported() {
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
	s.Require().ErrorIs(err, handlers.ErrUnsupportedTimeseriesDataType)
}

func (s *QueryDataSuite) TestBool() {
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

func (s *QueryDataSuite) TestMultipleFields() {
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

func (s *QueryDataSuite) TestMultipleFieldsWithNil() {
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

func (s *QueryDataSuite) TestDoNotRenderIntervals() {
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

func (s *QueryDataSuite) expectGetAndReturnTimeseries(q dataQuery, ts *telemetryapi.Timeseries) {
	s.mockTelemetryAPIClient.ExpectGetAndReturn(q.toGetParams(), ts, nil)
}

func (s *QueryDataSuite) expectGetAndReturnError(q dataQuery, err error) {
	s.mockTelemetryAPIClient.ExpectGetAndReturn(q.toGetParams(), nil, err)
}

func (s *QueryDataSuite) randomDataQuery() dataQuery {
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

func (s *QueryDataSuite) handleQuery(q dataQuery) (data.Frames, error) {
	var user *backend.User
	if q.user != "" {
		user = &backend.User{
			Email: q.user,
		}
	}

	pCtx := backend.PluginContext{
		User: user,
	}
	dataQuery := backend.DataQuery{
		TimeRange: backend.TimeRange{From: q.from, To: q.to},
		Interval:  q.interval,
		JSON: s.shouldMarshalJSON(map[string]interface{}{
			"text": q.text,
			"hide": q.hide,
		}),
	}
	return s.queryHandler.HandleQuery(s.ctx, pCtx, dataQuery)
}

func (s *QueryDataSuite) shouldMarshalJSON(o interface{}) []byte {
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
	hide     bool
}

func (q dataQuery) toGetParams() telemetryapi.TimeseriesParams {
	return telemetryapi.TimeseriesParams{
		User:  q.user,
		Query: q.text,
		From:  q.from,
		To:    q.to,
	}
}

func TestQueryData(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(QueryDataSuite))
}
