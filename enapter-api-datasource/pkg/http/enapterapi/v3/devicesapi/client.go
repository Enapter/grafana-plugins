package devicesapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi"
)

type ClientParams struct {
	HTTPClient *http.Client
	BaseURL    string
	Token      string
}

const DefaultTimeout = 15 * time.Second

func NewClient(p ClientParams) *Client {
	if p.HTTPClient == nil {
		p.HTTPClient = &http.Client{
			Timeout: DefaultTimeout,
		}
	}
	if p.BaseURL == "" {
		panic("BaseURL missing or empty")
	}
	return &Client{
		httpClient: p.HTTPClient,
		baseURL:    p.BaseURL,
		token:      p.Token,
	}
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

func (c *Client) Close() {
	c.httpClient.CloseIdleConnections()
}

func (c *Client) Ready(ctx context.Context) error {
	_, err := c.GetManifest(ctx, GetManifestParams{
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
	// FIXME: Cannot differentiate between a non-existing API endpoint and
	// a non-existing device as there's no error code in the response at
	// the moment. Checking the text of the error message could help but
	// that is unreliable as message are subject to change.
	return nil
}

type GetManifestParams struct {
	User     string
	DeviceID string
}

func (c *Client) GetManifest(
	ctx context.Context, p GetManifestParams,
) (_ []byte, retErr error) {
	req, err := c.newGetManifestRequest(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() {
		if err := enapterapi.DrainAndClose(resp.Body); err != nil {
			if retErr == nil {
				retErr = err
			}
		}
	}()

	manifest, err := c.processGetManifestResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("process response: %w", err)
	}

	return manifest, nil
}

func (c *Client) newGetManifestRequest(
	ctx context.Context, p GetManifestParams,
) (*http.Request, error) {
	urlString := c.baseURL + fmt.Sprintf("/%s/manifest", p.DeviceID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlString, nil)
	if err != nil {
		return nil, err
	}

	req.Header["Accept"] = []string{"application/json"}

	if p.User != "" {
		const userField = "X-Enapter-Auth-User"
		req.Header[userField] = []string{p.User}
	}

	const tokenField = "X-Enapter-Auth-Token" //nolint: gosec // false positive
	req.Header[tokenField] = []string{c.token}

	return req, nil
}

func (c *Client) processGetManifestResponse(resp *http.Response) ([]byte, error) {
	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden,
		http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity,
		http.StatusTooManyRequests, http.StatusInternalServerError:
		return nil, c.processError(resp)
	default:
		return nil, c.processUnexpectedStatus(resp)
	}

	const wantContentType = "application/json"
	if have := resp.Header.Get("Content-Type"); have != wantContentType {
		return nil, fmt.Errorf("%w: want %s, have %s",
			errUnexpectedContentType, wantContentType, have)
	}

	var payload struct {
		Manifest json.RawMessage `json:"manifest"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}

	return payload.Manifest, nil
}

type ExecuteCommandParams struct {
	User     string
	DeviceID string
	Request  CommandRequest
}

type CommandRequest struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type CommandResponse struct {
	State   string         `json:"state"`
	Payload map[string]any `json:"payload"`
}

type CommandExecution struct {
	State    string          `json:"state"`
	Response CommandResponse `json:"response"`
}

func (c *Client) ExecuteCommand(
	ctx context.Context, p ExecuteCommandParams,
) (_ *CommandExecution, retErr error) {
	req, err := c.newExecuteCommandRequest(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() {
		if err := enapterapi.DrainAndClose(resp.Body); err != nil {
			if retErr == nil {
				retErr = err
			}
		}
	}()

	execution, err := c.processExecuteCommandResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("process response: %w", err)
	}

	return execution, nil
}

func (c *Client) newExecuteCommandRequest(
	ctx context.Context, p ExecuteCommandParams,
) (*http.Request, error) {
	urlString := c.baseURL + fmt.Sprintf("/%s/execute_command", p.DeviceID)

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(p.Request); err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlString, body)
	if err != nil {
		return nil, err
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.Header["Accept"] = []string{"application/json"}

	if p.User != "" {
		const userField = "X-Enapter-Auth-User"
		req.Header[userField] = []string{p.User}
	}

	const tokenField = "X-Enapter-Auth-Token" //nolint: gosec // false positive
	req.Header[tokenField] = []string{c.token}

	return req, nil
}

func (c *Client) processExecuteCommandResponse(
	resp *http.Response,
) (*CommandExecution, error) {
	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden,
		http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity,
		http.StatusTooManyRequests, http.StatusInternalServerError:
		return nil, c.processError(resp)
	default:
		return nil, c.processUnexpectedStatus(resp)
	}

	const wantContentType = "application/json"
	if have := resp.Header.Get("Content-Type"); have != wantContentType {
		return nil, fmt.Errorf("%w: want %s, have %s",
			errUnexpectedContentType, wantContentType, have)
	}

	var payload struct {
		Execution CommandExecution `json:"execution"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}

	return &payload.Execution, nil
}

func (c *Client) processError(resp *http.Response) error {
	multiErr, err := enapterapi.ParseMultiError(resp.Body)
	if err != nil {
		return fmt.Errorf("multi-error: <not available>: %w", err)
	}

	return multiErr
}

func (c *Client) processUnexpectedStatus(resp *http.Response) error {
	dump, err := enapterapi.DumpBody(resp.Body)
	if err != nil {
		//nolint:errorlint // two errors
		return fmt.Errorf("%w: %s: body dump: <not available>: %v",
			errUnexpectedStatus, resp.Status, err)
	}

	return fmt.Errorf("%w: %s: body dump: %s",
		errUnexpectedStatus, resp.Status, dump)
}
