package handlers_test

import (
	"context"

	"github.com/stretchr/testify/suite"

	"github.com/Enapter/grafana-plugins/pkg/telemetryapi"
)

var _ telemetryapi.Client = (*MockTelemetryAPIClient)(nil)

type MockTelemetryAPIClient struct {
	suite             *suite.Suite
	timeseriesHandler func(telemetryapi.TimeseriesParams) (
		*telemetryapi.Timeseries, error)
}

func NewMockTelemetryAPIClient(s *suite.Suite) *MockTelemetryAPIClient {
	c := new(MockTelemetryAPIClient)
	c.suite = s
	c.timeseriesHandler = c.unexpectedCall
	return c
}

func (c *MockTelemetryAPIClient) ExpectGetAndReturn(
	wantP telemetryapi.TimeseriesParams,
	timeseries *telemetryapi.Timeseries, err error,
) {
	c.timeseriesHandler = func(haveP telemetryapi.TimeseriesParams) (
		*telemetryapi.Timeseries, error,
	) {
		defer func() { c.timeseriesHandler = c.unexpectedCall }()
		c.suite.Require().Equal(wantP, haveP)
		return timeseries, err
	}
}

func (c *MockTelemetryAPIClient) Timeseries(
	_ context.Context, p telemetryapi.TimeseriesParams,
) (*telemetryapi.Timeseries, error) {
	return c.timeseriesHandler(p)
}

func (c *MockTelemetryAPIClient) unexpectedCall(telemetryapi.TimeseriesParams) (
	*telemetryapi.Timeseries, error,
) {
	c.suite.Require().FailNow("unexpected call")
	//nolint: nilnil // unreachable
	return nil, nil
}

func (c *MockTelemetryAPIClient) Ready(context.Context) error { return nil }
func (c *MockTelemetryAPIClient) Close()                      {}
