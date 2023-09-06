package handlers_test

import (
	"context"

	"github.com/stretchr/testify/suite"

	"github.com/Enapter/telemetry-grafana-datasource-plugin/pkg/commandsapi"
)

var _ commandsapi.Client = (*MockCommandsAPIClient)(nil)

type MockCommandsAPIClient struct {
	suite          *suite.Suite
	executeHandler func(commandsapi.ExecuteParams) (
		commandsapi.CommandResponse, error)
}

func NewMockCommandsAPIClient(s *suite.Suite) *MockCommandsAPIClient {
	c := new(MockCommandsAPIClient)
	c.suite = s
	c.executeHandler = c.unexpectedCall
	return c
}

func (c *MockCommandsAPIClient) ExpectExecuteAndReturn(
	wantP commandsapi.ExecuteParams,
	cmdResp commandsapi.CommandResponse, err error,
) {
	c.executeHandler = func(haveP commandsapi.ExecuteParams) (
		commandsapi.CommandResponse, error,
	) {
		defer func() { c.executeHandler = c.unexpectedCall }()
		c.suite.Require().Equal(wantP, haveP)
		return cmdResp, err
	}
}

func (c *MockCommandsAPIClient) Execute(
	_ context.Context, p commandsapi.ExecuteParams,
) (commandsapi.CommandResponse, error) {
	return c.executeHandler(p)
}

func (c *MockCommandsAPIClient) unexpectedCall(commandsapi.ExecuteParams) (
	commandsapi.CommandResponse, error,
) {
	c.suite.Require().FailNow("unexpected call")
	//nolint: nilnil // unreachable
	return commandsapi.CommandResponse{}, nil
}

func (c *MockCommandsAPIClient) Ready(context.Context) error { return nil }
func (c *MockCommandsAPIClient) Close()                      {}
