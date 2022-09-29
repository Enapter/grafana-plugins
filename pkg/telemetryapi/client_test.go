package telemetryapi_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/suite"

	"github.com/Enapter/telemetry-grafana-datasource-plugin/pkg/telemetryapi"
)

type ClientSuite struct {
	suite.Suite
	ctx    context.Context
	token  string
	server *MockServer
	client telemetryapi.Client
}

func (s *ClientSuite) SetupTest() {
	s.ctx = context.Background()
	s.token = faker.Word()
	s.server = StartMockServer(s.T())
	client, err := telemetryapi.NewClient(telemetryapi.ClientParams{
		Logger:     hclog.New(hclog.DefaultOptions),
		HTTPClient: s.server.NewClient(),
		BaseURL:    s.server.Address(),
		Token:      s.token,
	})
	s.Require().NoError(err)
	s.client = client
}

func (s *ClientSuite) TearDownTest() {
	s.client.Close()
	s.server.Stop()
}

func (s *ClientSuite) TestReadyOK() {
	const errorJSON = `{"errors":[{"code":"unprocessable_entity","message":"Oops."}]}`
	s.server.ExpectTimeseriesRequestAndReturnCode(
		http.StatusUnprocessableEntity, errorJSON)

	err := s.client.Ready(s.ctx)
	s.Require().NoError(err)
}

func (s *ClientSuite) TestReadyUnexpectedCode() {
	const errorJSON = `{"errors":[{"code":"invalid_query_parameter_format","message":"Oops."}]}`
	s.server.ExpectTimeseriesRequestAndReturnCode(
		http.StatusUnprocessableEntity, errorJSON)
	err := s.client.Ready(s.ctx)
	s.Require().Error(err)
	s.Require().Equal(
		`process timeseries response: code=invalid_query_parameter_format, message="Oops."`,
		err.Error())
}

func (s *ClientSuite) TestReadyUnexpectedAbsenseOfError() {
	s.server.ExpectTimeseriesRequestAndReturnData([]string{"float64"}, `
ts,k=v
1,2.1
3,4.2
5,6.3
`)
	err := s.client.Ready(s.ctx)
	s.Require().Error(err)
	s.Require().Equal(
		"unexpected absence of error",
		err.Error())
}

func (s *ClientSuite) TestGetZeroContentLength() {
	s.server.ExpectTimeseriesRequestAndReturnZeroContentLength()
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().Error(err)
	s.Require().Nil(timeseries)
	s.Require().Equal(
		"process timeseries response: no values",
		err.Error())
}

func (s *ClientSuite) TestGetEmptyUser() {
	p := s.randomGetParams()
	p.User = ""
	s.server.ExpectTimeseriesRequestAndReturnData([]string{"bool"}, `
ts,k=v
1,true
`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().NoError(err)
	s.Require().Equal(timeseries.TimeField[0].Unix(), int64(1))
	s.Require().Equal(*timeseries.DataFields[0].Values[0].(*bool), true)
}

func (s *ClientSuite) TestGetInvalidContentType() {
	s.server.ExpectTimeseriesRequestAndReturnInvalidContentType()
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().Error(err)
	s.Require().Nil(timeseries)
	s.Require().Equal(
		"process timeseries response: unexpected content type: "+
			"want text/csv, have application/json",
		err.Error())
}

func (s *ClientSuite) TestGetEmptyDataTypes() {
	s.server.ExpectTimeseriesRequestAndReturnData([]string{}, `hello,anyone?`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().Error(err)
	s.Require().Nil(timeseries)
	s.Require().Equal(
		"process timeseries response: empty header field: "+
			"X-Enapter-Timeseries-Data-Types",
		err.Error())
}

func (s *ClientSuite) TestGetInvalidDataTypes() {
	s.server.ExpectTimeseriesRequestAndReturnData([]string{"invalid"}, `foo,bar`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().Error(err)
	s.Require().Nil(timeseries)
	s.Require().Equal(
		"process timeseries response: parse data types: 0: "+
			"unexpected timeseries data type: invalid",
		err.Error())
}

func (s *ClientSuite) TestGetMultiError() {
	for _, code := range []int{
		http.StatusBadRequest,
		http.StatusForbidden,
		http.StatusUnprocessableEntity,
		http.StatusTooManyRequests,
	} {
		s.Run(strconv.Itoa(code), func() {
			s.server.ExpectTimeseriesRequestAndReturnCode(code,
				`{"errors":[{"code":"oops"}]}`)
			timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
			s.Require().Error(err)
			s.Require().Nil(timeseries)
			s.Require().Equal(
				"process timeseries response: code=oops",
				err.Error())
		})
	}
}

func (s *ClientSuite) TestGetMultiErrorEmptyErrors() {
	s.server.ExpectTimeseriesRequestAndReturnCode(http.StatusBadRequest,
		`{"errors":[]}`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().Error(err)
	s.Require().Nil(timeseries)
	s.Require().Equal(
		"process timeseries response: multi-error: <not available>: empty error list",
		err.Error())
}

func (s *ClientSuite) TestGetMultiErrorEmptyCode() {
	s.server.ExpectTimeseriesRequestAndReturnCode(http.StatusBadRequest,
		`{"errors":[{"message":"oops"}]}`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().Error(err)
	s.Require().Nil(timeseries)
	s.Require().Equal(
		`process timeseries response: code=<empty>, message="oops"`,
		err.Error())
}

func (s *ClientSuite) TestGetUnexpectedStatusWithNoDescription() {
	s.server.ExpectTimeseriesRequestAndReturnCode(http.StatusTeapot, ``)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().Error(err)
	s.Require().Nil(timeseries)
	s.Require().Equal(
		"process timeseries response: unexpected status: 418 I'm a teapot: "+
			"body dump: <not available>: empty data",
		err.Error())
}

func (s *ClientSuite) TestGetEmptyRecords() {
	s.server.ExpectTimeseriesRequestAndReturnData([]string{"float64"}, ``)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().ErrorIs(err, telemetryapi.ErrNoValues)
	s.Require().Nil(timeseries)
}

func (s *ClientSuite) TestGetRequestParams() {
	const user = "gizmo@enapter.com"
	const query = "DROP TABLE datapoints;"
	p := telemetryapi.TimeseriesParams{
		User:  user,
		Query: query,
	}
	checkFn := func(r *http.Request) {
		s.Require().Equal([]string{user}, r.Header["X-Enapter-Auth-User"])
		s.Require().Equal([]string{s.token}, r.Header["X-Enapter-Auth-Token"])
		data, err := ioutil.ReadAll(r.Body)
		s.Require().NoError(err)
		s.Require().NoError(r.Body.Close())
		r.Body = ioutil.NopCloser(bytes.NewReader(data))
		s.Require().Equal(query, string(data))
	}
	s.server.ExpectTimeseriesRequestCheckItAndReturnData(checkFn, "float64", `
ts,k=v
1,2.1
3,4.0
5,6.3
`)
	timeseries, err := s.client.Timeseries(s.ctx, p)
	s.Require().NoError(err)
	s.Require().Positive(timeseries.Len())
}

func (s *ClientSuite) TestInvalidCSVHeader() {
	headers := []string{"ts", "non_ts,k=v"}
	for _, h := range headers {
		s.Run(h, func() {
			s.server.ExpectTimeseriesRequestAndReturnData([]string{"float64"}, h)
			_, err := s.client.Timeseries(s.ctx, s.randomGetParams())
			s.Require().Error(err)
		})
	}
}

func (s *ClientSuite) TestGetMultipleFields() {
	s.server.ExpectTimeseriesRequestAndReturnData([]string{"float64", "bool"}, `
ts,k1=v1,k2=v2
1,2.1,true
3,4.2,false
5,6.3,true
`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().NoError(err)
	s.Require().Equal(timeseries.TimeField[0].Unix(), int64(1))
	s.Require().Equal(timeseries.TimeField[1].Unix(), int64(3))
	s.Require().Equal(timeseries.TimeField[2].Unix(), int64(5))
	s.Require().Equal(*timeseries.DataFields[0].Values[0].(*float64), 2.1)
	s.Require().Equal(*timeseries.DataFields[0].Values[1].(*float64), 4.2)
	s.Require().Equal(*timeseries.DataFields[0].Values[2].(*float64), 6.3)
	s.Require().Equal(*timeseries.DataFields[1].Values[0].(*bool), true)
	s.Require().Equal(*timeseries.DataFields[1].Values[1].(*bool), false)
	s.Require().Equal(*timeseries.DataFields[1].Values[2].(*bool), true)
}

func (s *ClientSuite) TestGetMultipleFieldsWithNil() {
	s.server.ExpectTimeseriesRequestAndReturnData([]string{"float64", "bool"}, `
ts,k1=v1,k2=v2
1,2.1,true
3,,false
5,6.3,true
`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().NoError(err)
	s.Require().Equal(timeseries.TimeField[0].Unix(), int64(1))
	s.Require().Equal(timeseries.TimeField[1].Unix(), int64(3))
	s.Require().Equal(timeseries.TimeField[2].Unix(), int64(5))
	s.Require().Equal(*timeseries.DataFields[0].Values[0].(*float64), 2.1)
	s.Require().Equal(timeseries.DataFields[0].Values[1].(*float64), (*float64)(nil))
	s.Require().Equal(*timeseries.DataFields[0].Values[2].(*float64), 6.3)
	s.Require().Equal(*timeseries.DataFields[1].Values[0].(*bool), true)
	s.Require().Equal(*timeseries.DataFields[1].Values[1].(*bool), false)
	s.Require().Equal(*timeseries.DataFields[1].Values[2].(*bool), true)
}

func (s *ClientSuite) TestGetWrongNumberOfFields() {
	s.server.ExpectTimeseriesRequestAndReturnData([]string{"float64"}, `
ts,k=v
1,2.1
3
5,6.3
`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().Error(err)
	s.Require().Equal(
		"process timeseries response: parse CSV: parse record 2: "+
			"unexpected number of fields: want 2, have 1",
		err.Error())
	s.Require().Nil(timeseries)
}

//nolint: dupl // FIXME
func (s *ClientSuite) TestGetFloatRecords() {
	s.server.ExpectTimeseriesRequestAndReturnData([]string{"float64"}, `
ts,k=v
1,2.1
3,4.2
5,6.3
`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().NoError(err)
	s.Require().Equal(timeseries.TimeField[0].Unix(), int64(1))
	s.Require().Equal(timeseries.TimeField[1].Unix(), int64(3))
	s.Require().Equal(timeseries.TimeField[2].Unix(), int64(5))
	s.Require().Equal(*timeseries.DataFields[0].Values[0].(*float64), 2.1)
	s.Require().Equal(*timeseries.DataFields[0].Values[1].(*float64), 4.2)
	s.Require().Equal(*timeseries.DataFields[0].Values[2].(*float64), 6.3)
}

func (s *ClientSuite) TestGetIntRecords() {
	s.server.ExpectTimeseriesRequestAndReturnData([]string{"int64"}, `
ts,k=v
11,22
33,44
55,66
`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().NoError(err)
	s.Require().Equal(timeseries.TimeField[0].Unix(), int64(11))
	s.Require().Equal(timeseries.TimeField[1].Unix(), int64(33))
	s.Require().Equal(timeseries.TimeField[2].Unix(), int64(55))
	s.Require().Equal(*timeseries.DataFields[0].Values[0].(*int64), int64(22))
	s.Require().Equal(*timeseries.DataFields[0].Values[1].(*int64), int64(44))
	s.Require().Equal(*timeseries.DataFields[0].Values[2].(*int64), int64(66))
}

//nolint: dupl // FIXME
func (s *ClientSuite) TestGetStringRecords() {
	s.server.ExpectTimeseriesRequestAndReturnData([]string{"string"}, `
ts,k=v
1,foo
2,bar
3,baz
`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().NoError(err)
	s.Require().Equal(timeseries.TimeField[0].Unix(), int64(1))
	s.Require().Equal(timeseries.TimeField[1].Unix(), int64(2))
	s.Require().Equal(timeseries.TimeField[2].Unix(), int64(3))
	s.Require().Equal(*timeseries.DataFields[0].Values[0].(*string), "foo")
	s.Require().Equal(*timeseries.DataFields[0].Values[1].(*string), "bar")
	s.Require().Equal(*timeseries.DataFields[0].Values[2].(*string), "baz")
}

func (s *ClientSuite) TestGetBooleanRecords() {
	s.server.ExpectTimeseriesRequestAndReturnData([]string{"bool"}, `
ts,k=v
1,true
2,true
3,false
`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().NoError(err)
	s.Require().Equal(timeseries.TimeField[0].Unix(), int64(1))
	s.Require().Equal(timeseries.TimeField[1].Unix(), int64(2))
	s.Require().Equal(timeseries.TimeField[2].Unix(), int64(3))
	s.Require().Equal(*timeseries.DataFields[0].Values[0].(*bool), true)
	s.Require().Equal(*timeseries.DataFields[0].Values[1].(*bool), true)
	s.Require().Equal(*timeseries.DataFields[0].Values[2].(*bool), false)
}

func (s *ClientSuite) randomGetParams() telemetryapi.TimeseriesParams {
	return telemetryapi.TimeseriesParams{
		User:  faker.Email(),
		Query: faker.Sentence(),
	}
}

func TestClient(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ClientSuite))
}
