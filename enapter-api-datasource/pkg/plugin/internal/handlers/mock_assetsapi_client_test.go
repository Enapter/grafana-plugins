package handlers_test

import (
	"context"

	"github.com/stretchr/testify/suite"

	"github.com/Enapter/grafana-plugins/pkg/assetsapi"
)

type MockAssetsAPIClient struct {
	suite             *suite.Suite
	deviceByIDHandler func(assetsapi.DeviceByIDParams) (*assetsapi.Device, error)
}

func NewMockAssetsAPIClient(s *suite.Suite) *MockAssetsAPIClient {
	c := new(MockAssetsAPIClient)
	c.suite = s
	c.deviceByIDHandler = c.unexpectedCall
	return c
}

func (c *MockAssetsAPIClient) ExpectDeviceByIDAndReturn(
	wantP assetsapi.DeviceByIDParams,
	device *assetsapi.Device, err error,
) {
	c.deviceByIDHandler = func(haveP assetsapi.DeviceByIDParams) (
		*assetsapi.Device, error,
	) {
		defer func() { c.deviceByIDHandler = c.unexpectedCall }()
		c.suite.Require().Equal(wantP, haveP)
		return device, err
	}
}

func (c *MockAssetsAPIClient) DeviceByID(
	_ context.Context, p assetsapi.DeviceByIDParams,
) (*assetsapi.Device, error) {
	return c.deviceByIDHandler(p)
}

func (c *MockAssetsAPIClient) unexpectedCall(assetsapi.DeviceByIDParams) (
	*assetsapi.Device, error,
) {
	c.suite.Require().FailNow("unexpected call")
	//nolint: nilnil // unreachable
	return nil, nil
}

func (c *MockAssetsAPIClient) Ready(context.Context) error { return nil }
func (c *MockAssetsAPIClient) Close()                      {}
