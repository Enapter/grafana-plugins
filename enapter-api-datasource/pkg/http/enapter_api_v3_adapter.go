package http

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-hclog"

	"github.com/Enapter/grafana-plugins/pkg/core"
	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi"
	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi/v3/devicesapi"
	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi/v3/telemetryapi"
)

type EnapterAPIv3AdapterParams struct {
	Logger   hclog.Logger
	APIURL   string
	APIToken string
}

type EnapterAPIv3Adapter struct {
	logger             hclog.Logger
	telemetryAPIClient *telemetryapi.Client
	devicesAPIClient   *devicesapi.Client
}

func NewEnapterAPIv3Adapter(p EnapterAPIv3AdapterParams) *EnapterAPIv3Adapter {
	telemetryAPIClient := telemetryapi.NewClient(telemetryapi.ClientParams{
		BaseURL: p.APIURL + "/v3/telemetry",
		Token:   p.APIToken,
	})
	devicesAPIClient := devicesapi.NewClient(devicesapi.ClientParams{
		BaseURL: p.APIURL + "/v3/devices",
		Token:   p.APIToken,
	})
	return &EnapterAPIv3Adapter{
		logger:             p.Logger.Named("enapter_api_v3_adapter"),
		telemetryAPIClient: telemetryAPIClient,
		devicesAPIClient:   devicesAPIClient,
	}
}

func (a *EnapterAPIv3Adapter) Close() {
	a.telemetryAPIClient.Close()
}

func (a *EnapterAPIv3Adapter) Ready(ctx context.Context) error {
	if err := a.telemetryAPIClient.Ready(ctx); err != nil {
		return fmt.Errorf("telemetry api: %w", err)
	}
	if err := a.devicesAPIClient.Ready(ctx); err != nil {
		return fmt.Errorf("devices api: %w", err)
	}
	return nil
}

func (a *EnapterAPIv3Adapter) QueryTimeseries(
	ctx context.Context, req *core.QueryTimeseriesRequest,
) (*core.QueryTimeseriesResponse, error) {
	timeseries, err := a.telemetryAPIClient.Timeseries(
		ctx, telemetryapi.TimeseriesParams{
			User:  req.User,
			Query: req.Query,
		})
	if err != nil {
		if errors.Is(err, telemetryapi.ErrNoValues) {
			return nil, core.ErrTimeseriesEmpty
		}
		if multiErr := new(enapterapi.MultiError); errors.As(err, &multiErr) {
			return nil, a.convertMultiError(multiErr)
		}
		return nil, err
	}
	dataFields := make([]*core.TimeseriesDataField, len(timeseries.DataFields))
	for i, field := range timeseries.DataFields {
		dataFields[i] = &core.TimeseriesDataField{
			Tags:   core.TimeseriesTags(field.Tags),
			Type:   core.TimeseriesDataType(field.Type),
			Values: field.Values,
		}
	}
	return &core.QueryTimeseriesResponse{
		Timeseries: &core.Timeseries{
			TimeField:  timeseries.TimeField,
			DataFields: dataFields,
		},
	}, nil
}

func (a *EnapterAPIv3Adapter) ExecuteCommand(
	ctx context.Context, req *core.ExecuteCommandRequest,
) (*core.ExecuteCommandResponse, error) {
	cmdResp, err := a.devicesAPIClient.ExecuteCommand(ctx,
		devicesapi.ExecuteCommandParams{
			User:     req.User,
			DeviceID: req.DeviceID,
			Request: devicesapi.CommandRequest{
				Name:      req.CommandName,
				Arguments: req.CommandArgs,
			},
		})
	if err != nil {
		if multiErr := new(enapterapi.MultiError); errors.As(err, &multiErr) {
			return nil, a.convertMultiError(multiErr)
		}
		return nil, err
	}
	return &core.ExecuteCommandResponse{
		State:   strings.ToLower(cmdResp.Response.State),
		Payload: cmdResp.Response.Payload,
	}, nil
}

func (a *EnapterAPIv3Adapter) GetDeviceManifest(
	ctx context.Context, req *core.GetDeviceManifestRequest,
) (*core.GetDeviceManifestResponse, error) {
	manifest, err := a.devicesAPIClient.GetManifest(ctx, devicesapi.GetManifestParams{
		User:     req.User,
		DeviceID: req.DeviceID,
	})
	if err != nil {
		if multiErr := new(enapterapi.MultiError); errors.As(err, &multiErr) {
			return nil, a.convertMultiError(multiErr)
		}
		return nil, err
	}
	return &core.GetDeviceManifestResponse{
		Manifest: manifest,
	}, nil
}

func (a *EnapterAPIv3Adapter) convertMultiError(
	multiErr *enapterapi.MultiError,
) error {
	switch len(multiErr.Errors) {
	case 0:
		// should never happen, return error as is
		return multiErr
	case 1:
		return core.EnapterAPIError{
			Code:    multiErr.Errors[0].Code,
			Message: multiErr.Errors[0].Message,
			Details: multiErr.Errors[0].Details,
		}
	default:
		a.logger.Warn("multi error contains multiple errors, " +
			"but this is not supported yet; will return only the first error")
		return core.EnapterAPIError{
			Code:    multiErr.Errors[0].Code,
			Message: multiErr.Errors[0].Message,
			Details: multiErr.Errors[0].Details,
		}
	}
}
