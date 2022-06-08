package telemetryapi_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/suite"

	"github.com/Enapter/grafana-plugins/telemetry-datasource/pkg/telemetryapi"
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
	s.server.ExpectTimeseriesRequestAndReturnData("float64", `
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
	_, err := s.client.Timeseries(s.ctx, p)
	s.Require().ErrorIs(err, telemetryapi.ErrEmptyUser)
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
	s.server.ExpectTimeseriesRequestAndReturnData("", `hello,anyone?`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().Error(err)
	s.Require().Nil(timeseries)
	s.Require().Equal(
		"process timeseries response: empty header field: "+
			"X-Enapter-Timeseries-Data-Types",
		err.Error())
}

func (s *ClientSuite) TestGetInvalidDataTypes() {
	s.server.ExpectTimeseriesRequestAndReturnData("invalid", `foo,bar`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().Error(err)
	s.Require().Nil(timeseries)
	s.Require().Equal(
		"process timeseries response: parse data type: "+
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
	s.server.ExpectTimeseriesRequestAndReturnData("float64", ``)
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
		From:  time.Date(2020, time.November, 6, 5, 4, 3, 0, time.UTC),
		To:    time.Date(2021, time.December, 11, 22, 33, 44, 0, time.UTC),
	}
	checkFn := func(r *http.Request) {
		s.Require().Equal([]string{user}, r.Header["X-Enapter-Auth-User"])
		s.Require().Equal([]string{s.token}, r.Header["X-Enapter-Auth-Token"])
		q, err := url.ParseQuery(r.URL.RawQuery)
		s.Require().NoError(err)
		s.Require().Equal([]string{"2020-11-06T05:04:03Z"}, q["from"])
		s.Require().Equal([]string{"2021-12-11T22:33:44Z"}, q["to"])
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
	s.Require().NotEmpty(timeseries.Values)
}

func (s *ClientSuite) TestInvalidCSVHeader() {
	headers := []string{"ts", "non_ts,k=v"}
	for _, h := range headers {
		s.Run(h, func() {
			s.server.ExpectTimeseriesRequestAndReturnData("float64", h)
			_, err := s.client.Timeseries(s.ctx, s.randomGetParams())
			s.Require().Error(err)
		})
	}
}

func (s *ClientSuite) TestGetWrongNumberOfFields() {
	s.server.ExpectTimeseriesRequestAndReturnData("float64", `
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
	s.server.ExpectTimeseriesRequestAndReturnData("float64", `
ts,k=v
1,2.1
3,4.2
5,6.3
`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().NoError(err)
	s.Require().Equal(timeseries.Values[0].Timestamp.Unix(), int64(1))
	s.Require().Equal(timeseries.Values[1].Timestamp.Unix(), int64(3))
	s.Require().Equal(timeseries.Values[2].Timestamp.Unix(), int64(5))
	s.Require().Equal(timeseries.Values[0].Value.(float64), 2.1)
	s.Require().Equal(timeseries.Values[1].Value.(float64), 4.2)
	s.Require().Equal(timeseries.Values[2].Value.(float64), 6.3)
}

func (s *ClientSuite) TestGetIntRecords() {
	s.server.ExpectTimeseriesRequestAndReturnData("int64", `
ts,k=v
11,22
33,44
55,66
`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().NoError(err)
	s.Require().Equal(timeseries.Values[0].Timestamp.Unix(), int64(11))
	s.Require().Equal(timeseries.Values[1].Timestamp.Unix(), int64(33))
	s.Require().Equal(timeseries.Values[2].Timestamp.Unix(), int64(55))
	s.Require().Equal(timeseries.Values[0].Value.(int64), int64(22))
	s.Require().Equal(timeseries.Values[1].Value.(int64), int64(44))
	s.Require().Equal(timeseries.Values[2].Value.(int64), int64(66))
}

//nolint: dupl // FIXME
func (s *ClientSuite) TestGetStringRecords() {
	s.server.ExpectTimeseriesRequestAndReturnData("string", `
ts,k=v
1,foo
2,bar
3,baz
`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().NoError(err)
	s.Require().Equal(timeseries.Values[0].Timestamp.Unix(), int64(1))
	s.Require().Equal(timeseries.Values[1].Timestamp.Unix(), int64(2))
	s.Require().Equal(timeseries.Values[2].Timestamp.Unix(), int64(3))
	s.Require().Equal(timeseries.Values[0].Value.(string), "foo")
	s.Require().Equal(timeseries.Values[1].Value.(string), "bar")
	s.Require().Equal(timeseries.Values[2].Value.(string), "baz")
}

func (s *ClientSuite) TestGetBooleanRecords() {
	s.server.ExpectTimeseriesRequestAndReturnData("bool", `
ts,k=v
1,true
2,true
3,false
`)
	timeseries, err := s.client.Timeseries(s.ctx, s.randomGetParams())
	s.Require().NoError(err)
	s.Require().Equal(timeseries.Values[0].Timestamp.Unix(), int64(1))
	s.Require().Equal(timeseries.Values[1].Timestamp.Unix(), int64(2))
	s.Require().Equal(timeseries.Values[2].Timestamp.Unix(), int64(3))
	s.Require().Equal(timeseries.Values[0].Value.(bool), true)
	s.Require().Equal(timeseries.Values[1].Value.(bool), true)
	s.Require().Equal(timeseries.Values[2].Value.(bool), false)
}

func (s *ClientSuite) randomGetParams() telemetryapi.TimeseriesParams {
	to := time.Now().Add(-time.Duration(rand.Int()) * time.Second)
	from := to.Add(-time.Duration(rand.Int()) * time.Second)
	return telemetryapi.TimeseriesParams{
		User:  faker.Email(),
		Query: faker.Sentence(),
		From:  from,
		To:    to,
	}
}

func TestClient(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ClientSuite))
}
