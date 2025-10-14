package commandsapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	enapterhttp "github.com/Enapter/http-api-go-client/pkg/client"

	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi"
)

type Client struct {
	apiURL  string
	token   string
	timeout time.Duration
}

type ClientParams struct {
	APIURL  string
	Token   string
	Timeout time.Duration
}

const DefaultTimeout = 15 * time.Second

func NewClient(p ClientParams) *Client {
	if p.APIURL == "" {
		panic("APIURL missing or empty")
	}
	if p.Timeout == 0 {
		p.Timeout = DefaultTimeout
	}
	return &Client{
		apiURL:  p.APIURL,
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

func (c *Client) Execute(ctx context.Context, p ExecuteParams) (CommandResponse, error) {
	enapterHTTPClient, err := c.newEnapterHTTPClient(p.User)
	if err != nil {
		return CommandResponse{}, fmt.Errorf("new Enapter HTTP client: %w", err)
	}

	resp, err := enapterHTTPClient.Commands.Execute(ctx, enapterhttp.CommandQuery{
		DeviceID:    p.Request.DeviceID,
		HardwareID:  p.Request.HardwareID,
		CommandName: p.Request.CommandName,
		Arguments:   p.Request.CommandArgs,
	})
	if err != nil {
		if respErr := (enapterhttp.ResponseError{}); errors.As(err, &respErr) {
			return CommandResponse{}, c.respErrorToMultiError(respErr)
		}
		return CommandResponse{}, fmt.Errorf("do: %w", err)
	}

	return CommandResponse{
		State:   string(resp.State),
		Payload: resp.Payload,
	}, nil
}

func (c *Client) respErrorToMultiError(respErr enapterhttp.ResponseError) error {
	if len(respErr.Errors) == 0 {
		return respErr
	}

	multiErr := new(enapterapi.MultiError)

	for _, e := range respErr.Errors {
		if len(e.Code) == 0 {
			e.Code = "<empty>"
		}
		multiErr.Errors = append(multiErr.Errors, enapterapi.Error{
			Code:    e.Code,
			Message: e.Message,
			Details: e.Details,
		})
	}

	return multiErr
}

func (c *Client) newEnapterHTTPClient(user string) (*enapterhttp.Client, error) {
	transport := http.DefaultTransport

	if c.token != "" {
		transport = enapterhttp.NewAuthTokenTransport(transport, c.token)
	}

	if user != "" {
		transport = enapterhttp.NewAuthUserTransport(transport, user)
	}

	return enapterhttp.NewClientWithURL(&http.Client{
		Timeout:   c.timeout,
		Transport: transport,
	}, c.apiURL)
}
