package handlers_test

import (
	"context"

	"github.com/stretchr/testify/suite"

	"github.com/Enapter/grafana-plugins/pkg/core"
)

type MockEnapterAPIAdapter struct {
	suite                  *suite.Suite
	queryTimeseriesHandler func(
		context.Context, *core.QueryTimeseriesRequest,
	) (*core.QueryTimeseriesResponse, error)
	executeCommandHandler func(
		context.Context, *core.ExecuteCommandRequest,
	) (*core.ExecuteCommandResponse, error)
	getDeviceManifestHandler func(
		context.Context, *core.GetDeviceManifestRequest,
	) (*core.GetDeviceManifestResponse, error)
}

func NewMockEnapterAPIAdapter(s *suite.Suite) *MockEnapterAPIAdapter {
	c := new(MockEnapterAPIAdapter)
	c.suite = s
	c.queryTimeseriesHandler = c.unexpectedQueryTimeseriesCall
	c.executeCommandHandler = c.unexpectedExecuteCommandCall
	c.getDeviceManifestHandler = c.unexpectedGetDeviceManifestCall
	return c
}

func (c *MockEnapterAPIAdapter) ExpectQueryTimeseriesAndReturn(
	wantReq *core.QueryTimeseriesRequest,
	resp *core.QueryTimeseriesResponse, err error,
) {
	c.queryTimeseriesHandler = func(
		_ context.Context, haveReq *core.QueryTimeseriesRequest,
	) (*core.QueryTimeseriesResponse, error) {
		defer func() {
			c.queryTimeseriesHandler = c.unexpectedQueryTimeseriesCall
		}()
		c.suite.Require().Equal(wantReq, haveReq)
		return resp, err
	}
}

func (c *MockEnapterAPIAdapter) QueryTimeseries(
	ctx context.Context, req *core.QueryTimeseriesRequest,
) (*core.QueryTimeseriesResponse, error) {
	return c.queryTimeseriesHandler(ctx, req)
}

func (c *MockEnapterAPIAdapter) unexpectedQueryTimeseriesCall(
	context.Context, *core.QueryTimeseriesRequest,
) (*core.QueryTimeseriesResponse, error) {
	c.suite.Require().FailNow("unexpected call")
	//nolint: nilnil // unreachable
	return nil, nil
}

func (c *MockEnapterAPIAdapter) ExpectExecuteCommandAndReturn(
	wantReq *core.ExecuteCommandRequest,
	resp *core.ExecuteCommandResponse, err error,
) {
	c.executeCommandHandler = func(
		_ context.Context, haveReq *core.ExecuteCommandRequest,
	) (*core.ExecuteCommandResponse, error) {
		defer func() {
			c.executeCommandHandler = c.unexpectedExecuteCommandCall
		}()
		c.suite.Require().Equal(wantReq, haveReq)
		return resp, err
	}
}

func (c *MockEnapterAPIAdapter) ExecuteCommand(
	ctx context.Context, req *core.ExecuteCommandRequest,
) (*core.ExecuteCommandResponse, error) {
	return c.executeCommandHandler(ctx, req)
}

func (c *MockEnapterAPIAdapter) unexpectedExecuteCommandCall(
	context.Context, *core.ExecuteCommandRequest,
) (*core.ExecuteCommandResponse, error) {
	c.suite.Require().FailNow("unexpected call")
	//nolint: nilnil // unreachable
	return nil, nil
}

func (c *MockEnapterAPIAdapter) GetDeviceManifest(
	ctx context.Context, req *core.GetDeviceManifestRequest,
) (*core.GetDeviceManifestResponse, error) {
	return c.getDeviceManifestHandler(ctx, req)
}

func (c *MockEnapterAPIAdapter) unexpectedGetDeviceManifestCall(
	context.Context, *core.GetDeviceManifestRequest,
) (*core.GetDeviceManifestResponse, error) {
	c.suite.Require().FailNow("unexpected call")
	//nolint: nilnil // unreachable
	return nil, nil
}

func (c *MockEnapterAPIAdapter) Ready(context.Context) error { return nil }
