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

	"github.com/Enapter/telemetry-grafana-datasource-plugin/pkg/plugin/internal/handlers"
	"github.com/Enapter/telemetry-grafana-datasource-plugin/pkg/telemetryapi"
)

type QueryDataSuite struct {
	suite.Suite
	ctx                    context.Context
	logger                 hclog.Logger
	mockTelemetryAPIClient *MockTelemetryAPIClient
	queryDataHandler       *handlers.QueryData
}

func (s *QueryDataSuite) SetupSuite() {
	s.ctx = context.Background()
	s.mockTelemetryAPIClient = NewMockTelemetryAPIClient(s.Suite)
	s.logger = hclog.Default()
	s.queryDataHandler = handlers.NewQueryData(s.logger, s.mockTelemetryAPIClient)
}

var errFake = errors.New("fake error")

func (s *QueryDataSuite) TestTelemetryAPIError() {
	req := s.randomDataRequestWithSingleQuery()
	s.expectGetAndReturnError(req, errFake)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().ErrorIs(err, handlers.ErrSomethingWentWrong)
	s.Require().Nil(frames)
}

func (s *QueryDataSuite) TestHandleNoValuesError() {
	req := s.randomDataRequestWithSingleQuery()
	s.expectGetAndReturnError(req, telemetryapi.ErrNoValues)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().NoError(err)
	s.Require().Nil(frames)
}

func (s *QueryDataSuite) TestEmptyTextNoError() {
	req := s.randomDataRequestWithSingleQuery()
	req.queries[0].text = ""
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().NoError(err)
	s.Require().Nil(frames)
}

func (s *QueryDataSuite) TestInvalidYAML() {
	req := s.randomDataRequestWithSingleQuery()
	req.queries[0].text = "that's not yaml"
	timeseries := telemetryapi.NewTimeseries([]telemetryapi.TimeseriesDataType{
		telemetryapi.TimeseriesDataTypeInt64,
	})
	s.expectGetAndReturnTimeseries(req, timeseries)
	_, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().ErrorIs(err, handlers.ErrInvalidYAML)
}

func newFloat64(v float64) *float64 { return &v }
func newInt64(v int64) *int64       { return &v }
func newBool(v bool) *bool          { return &v }
func newString(v string) *string    { return &v }

//nolint: dupl // FIXME
func (s *QueryDataSuite) TestFloat64() {
	req := s.randomDataRequestWithSingleQuery()
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
	s.expectGetAndReturnTimeseries(req, timeseries)
	frames, err := s.handleDataRequestWithSingleQuery(req)
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
	req := s.randomDataRequestWithSingleQuery()
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
	s.expectGetAndReturnTimeseries(req, timeseries)
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
	req := s.randomDataRequestWithSingleQuery()
	req.queries[0].hide = true
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Nil(err)
	s.Require().Nil(frames)
}

func (s *QueryDataSuite) TestString() {
	req := s.randomDataRequestWithSingleQuery()
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
	s.expectGetAndReturnTimeseries(req, timeseries)
	frames, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Nil(err)
	timestampField, dataFields := s.extractTimeseriesFields(frames)
	s.Require().Len(dataFields, 1)
	s.Require().Equal(int64(1), timestampField.At(0).(time.Time).Unix())
	s.Require().Equal(int64(2), timestampField.At(1).(time.Time).Unix())
	s.Require().Equal("foo", *dataFields[0].At(0).(*string))
	s.Require().Equal("bar", *dataFields[0].At(1).(*string))
}

func (s *QueryDataSuite) TestNoUserInfo() {
	req := s.randomDataRequestWithSingleQuery()
	req.user = ""
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
	s.expectGetAndReturnTimeseries(req, timeseries)
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
	req := s.randomDataRequestWithSingleQuery()
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
	s.expectGetAndReturnTimeseries(req, timeseries)
	_, err := s.handleDataRequestWithSingleQuery(req)
	s.Require().Error(err, handlers.ErrMetricDataTypeIsNotSupported)
}

func (s *QueryDataSuite) TestBool() {
	req := s.randomDataRequestWithSingleQuery()
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
	s.expectGetAndReturnTimeseries(req, timeseries)
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
	req := s.randomDataRequestWithSingleQuery()
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
	s.expectGetAndReturnTimeseries(req, timeseries)
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
	req := s.randomDataRequestWithSingleQuery()
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
	s.expectGetAndReturnTimeseries(req, timeseries)
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
	req := s.randomDataRequestWithSingleQuery()
	req.queries[0].text = `{"A fact":"$__interval is $__interval_ms milliseconds."}`
	req.queries[0].interval = 42 * time.Second
	timeseries := telemetryapi.NewTimeseries([]telemetryapi.TimeseriesDataType{
		telemetryapi.TimeseriesDataTypeBool,
	})
	s.expectGetAndReturnTimeseries(req, timeseries)
	_, err := s.handleDataRequestWithSingleQuery(req)
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

func (s *QueryDataSuite) expectGetAndReturnTimeseries(req dataRequest, ts *telemetryapi.Timeseries) {
	for _, q := range req.queries {
		p := telemetryapi.TimeseriesParams{
			User:  req.user,
			Query: q.text,
			From:  q.from,
			To:    q.to,
		}
		s.mockTelemetryAPIClient.ExpectGetAndReturn(p, ts, nil)
	}
}

func (s *QueryDataSuite) expectGetAndReturnError(req dataRequest, err error) {
	for _, q := range req.queries {
		p := telemetryapi.TimeseriesParams{
			User:  req.user,
			Query: q.text,
			From:  q.from,
			To:    q.to,
		}
		s.mockTelemetryAPIClient.ExpectGetAndReturn(p, nil, err)
	}
}

func (s *QueryDataSuite) randomDataRequestWithSingleQuery() dataRequest {
	const abc = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	return dataRequest{
		user: faker.Email(),
		queries: []query{{
			refID:    string(abc[rand.Int()%len(abc)]),
			from:     time.Now().Add(-time.Duration(rand.Int()+1) * time.Hour),
			to:       time.Now().Add(-time.Duration(rand.Int()+1) * time.Minute),
			interval: time.Duration(rand.Int()) * time.Second,
			text: string(s.shouldMarshalJSON(map[string]interface{}{
				faker.Word(): faker.Sentence(),
				faker.Word(): rand.Int(),
				faker.Word(): rand.Int()%2 == 0,
			})),
		}},
	}
}

func (s *QueryDataSuite) handleDataRequestWithSingleQuery(req dataRequest) (data.Frames, error) {
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
			TimeRange: backend.TimeRange{From: q.from, To: q.to},
			Interval:  q.interval,
			JSON: s.shouldMarshalJSON(map[string]interface{}{
				"text": q.text,
				"hide": q.hide,
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
	refID    string
	from     time.Time
	to       time.Time
	interval time.Duration
	hide     bool
	text     string
}

func TestQueryData(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(QueryDataSuite))
}
