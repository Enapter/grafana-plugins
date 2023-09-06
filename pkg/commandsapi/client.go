package commandsapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	enapterhttp "github.com/Enapter/http-api-go-client/pkg/client"

	"github.com/Enapter/telemetry-grafana-datasource-plugin/pkg/httperr"
)

type Client interface {
	Execute(ctx context.Context, p ExecuteParams) (CommandResponse, error)
}

type client struct {
	token   string
	timeout time.Duration
}

type ClientParams struct {
	Token   string
	Timeout time.Duration
}

const DefaultTimeout = 15 * time.Second

func NewClient(p ClientParams) Client {
	if p.Timeout == 0 {
		p.Timeout = DefaultTimeout
	}

	return &client{
		token:   p.Token,
		timeout: p.Timeout,
	}
}

type ExecuteParams struct {
	User    string
	Request CommandRequest
}

type CommandRequest struct {
	CommandName string
	CommandArgs map[string]interface{}
	DeviceID    string
	HardwareID  string
}

type CommandResponse struct {
	State   string
	Payload map[string]interface{}
}

func (c *client) Execute(ctx context.Context, p ExecuteParams) (CommandResponse, error) {
	resp, err := c.enapterHTTP(p.User).Commands.Execute(ctx, enapterhttp.CommandQuery{
		DeviceID:    p.Request.DeviceID,
		HardwareID:  p.Request.HardwareID,
		CommandName: p.Request.CommandName,
		Arguments:   p.Request.CommandArgs,
	})
	if err != nil {
		if respErr := (enapterhttp.ResponseError{}); errors.As(err, &respErr) {
			return CommandResponse{}, c.respErrorToMultiError(respErr)
		}
		return CommandResponse{}, err
	}

	return CommandResponse{
		State:   string(resp.State),
		Payload: resp.Payload,
	}, nil
}

func (c *client) respErrorToMultiError(respErr enapterhttp.ResponseError) error {
	if len(respErr.Errors) == 0 {
		return respErr
	}

	multiErr := new(httperr.MultiError)

	for _, e := range respErr.Errors {
		if len(e.Code) == 0 {
			e.Code = "<empty>"
		}
		multiErr.Errors = append(multiErr.Errors, httperr.Error{
			Code:    e.Code,
			Message: e.Message,
			Details: e.Details,
		})
	}

	return multiErr
}

func (c *client) enapterHTTP(user string) *enapterhttp.Client {
	transport := http.DefaultTransport

	if c.token != "" {
		transport = enapterhttp.NewAuthTokenTransport(transport, c.token)
	}

	if user != "" {
		transport = enapterhttp.NewAuthUserTransport(transport, user)
	}

	return enapterhttp.NewClient(&http.Client{
		Timeout:   c.timeout,
		Transport: transport,
	})
}