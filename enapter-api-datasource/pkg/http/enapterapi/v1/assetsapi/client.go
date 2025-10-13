package assetsapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	enapterhttp "github.com/Enapter/http-api-go-client/pkg/client"

	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi"
)

type (
	Device             = enapterhttp.Device
	ExpandDeviceParams = enapterhttp.ExpandDeviceParams
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
	if p.Timeout == 0 {
		p.Timeout = DefaultTimeout
	}

	return &Client{
		apiURL:  p.APIURL,
		token:   p.Token,
		timeout: p.Timeout,
	}
}

type DeviceByIDParams struct {
	User     string
	DeviceID string
	Expand   ExpandDeviceParams
}

func (c *Client) Ready(ctx context.Context) error {
	_, err := c.DeviceByID(ctx, DeviceByIDParams{
		User:     "does_not_exist",
		DeviceID: "does_not_exist",
	})
	if err == nil {
		return errUnexpectedAbsenceOfError
	}

	var multiErr *enapterapi.MultiError
	if ok := errors.As(err, &multiErr); !ok || len(multiErr.Errors) != 1 {
		return err
	}

	const expectedCode = "not_found"
	if multiErr.Errors[0].Code != expectedCode {
		return err
	}

	return nil
}

func (c *Client) DeviceByID(
	ctx context.Context, p DeviceByIDParams,
) (*Device, error) {
	enapterHTTPClient, err := c.newEnapterHTTPClient(p.User)
	if err != nil {
		return nil, fmt.Errorf("new Enapter HTTP client: %w", err)
	}

	resp, err := enapterHTTPClient.Assets.DeviceByID(ctx, enapterhttp.DeviceByIDQuery{
		ID:     p.DeviceID,
		Expand: p.Expand,
	})
	if err != nil {
		if respErr := (enapterhttp.ResponseError{}); errors.As(err, &respErr) {
			return nil, c.respErrorToMultiError(respErr)
		}
		return nil, fmt.Errorf("do: %w", err)
	}

	return &resp.Device, nil
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
