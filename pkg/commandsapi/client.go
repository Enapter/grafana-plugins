package commandsapi

import (
	"context"
	"net/http"
	"time"

	enapterhttp "github.com/Enapter/http-api-go-client/pkg/client"
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
		return CommandResponse{}, err
	}

	return CommandResponse{
		State:   string(resp.State),
		Payload: resp.Payload,
	}, nil
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
