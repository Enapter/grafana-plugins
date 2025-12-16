package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Enapter/grafana-plugins/pkg/core"
	httputil "github.com/Enapter/grafana-plugins/pkg/http/util"
)

type UserResolverAdapterParams struct {
	URL     string
	Timeout time.Duration
}

type UserResolverAdapter struct {
	url        string
	httpClient http.Client
}

func NewUserResolverAdapter(p UserResolverAdapterParams) *UserResolverAdapter {
	if p.Timeout == 0 {
		p.Timeout = 15 * time.Second
	}
	return &UserResolverAdapter{
		httpClient: http.Client{
			Timeout: p.Timeout,
		},
		url: p.URL,
	}
}

func (a *UserResolverAdapter) ResolveUser(
	ctx context.Context, req *core.ResolveUserRequest,
) (_ *core.ResolveUserResponse, retErr error) {
	httpReq, err := a.newResolveUserRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	httpResp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() {
		if err := httputil.DrainAndClose(httpResp.Body); err != nil {
			if retErr == nil {
				retErr = err
			}
		}
	}()
	resp, err := a.processResolveUserResponse(httpResp)
	if err != nil {
		return nil, fmt.Errorf("process response: %w", err)
	}
	return resp, nil
}

func (a *UserResolverAdapter) newResolveUserRequest(
	ctx context.Context, req *core.ResolveUserRequest,
) (*http.Request, error) {
	urlString := a.url + "?" + url.Values{
		"username": []string{req.Email},
	}.Encode()
	return http.NewRequestWithContext(ctx, http.MethodGet, urlString, nil)
}

func (a *UserResolverAdapter) processResolveUserResponse(
	httpResp *http.Response,
) (*core.ResolveUserResponse, error) {
	if httpResp.StatusCode != http.StatusOK {
		return nil, a.processUnexpectedStatusCode(httpResp)
	}
	var payload struct {
		GUID string `json:"guid"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}
	return &core.ResolveUserResponse{
		ID: payload.GUID,
	}, nil
}

func (a *UserResolverAdapter) processUnexpectedStatusCode(resp *http.Response) error {
	dump, err := httputil.DumpBody(resp.Body)
	if err != nil {
		//nolint:errorlint // two errors
		return fmt.Errorf("%w: %s: body dump: <not available>: %v",
			errUnexpectedStatusCode, resp.Status, err)
	}
	return fmt.Errorf("%w: %s: body dump: %s",
		errUnexpectedStatusCode, resp.Status, dump)
}
