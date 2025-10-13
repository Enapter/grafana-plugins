package http

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-hclog"

	"github.com/Enapter/grafana-plugins/pkg/core"
	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi"
	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi/v1/assetsapi"
	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi/v1/commandsapi"
	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi/v1/telemetryapi"
)

type EnapterAPIv1AdapterParams struct {
	Logger   hclog.Logger
	APIURL   string
	APIToken string
}

type EnapterAPIv1Adapter struct {
	logger             hclog.Logger
	telemetryAPIClient *telemetryapi.Client
	commandsAPIClient  *commandsapi.Client
	assetsAPIClient    *assetsapi.Client
}

func NewEnapterAPIv1Adapter(p EnapterAPIv1AdapterParams) *EnapterAPIv1Adapter {
	telemetryAPIClient := telemetryapi.NewClient(telemetryapi.ClientParams{
		BaseURL: p.APIURL + "/telemetry",
		Token:   p.APIToken,
	})
	commandsAPIClient := commandsapi.NewClient(commandsapi.ClientParams{
		APIURL: p.APIURL,
		Token:  p.APIToken,
	})
	assetsAPIClient := assetsapi.NewClient(assetsapi.ClientParams{
		APIURL: p.APIURL,
		Token:  p.APIToken,
	})
	return &EnapterAPIv1Adapter{
		logger:             p.Logger.Named("enapter_api_v1_adapter"),
		telemetryAPIClient: telemetryAPIClient,
		commandsAPIClient:  commandsAPIClient,
		assetsAPIClient:    assetsAPIClient,
	}
}

func (a *EnapterAPIv1Adapter) Close() {
	a.telemetryAPIClient.Close()
}

func (a *EnapterAPIv1Adapter) Ready(ctx context.Context) error {
	if err := a.telemetryAPIClient.Ready(ctx); err != nil {
		return fmt.Errorf("telemetry api: %w", err)
	}
	if err := a.assetsAPIClient.Ready(ctx); err != nil {
		return fmt.Errorf("assets api: %w", err)
	}
	return nil
}

func (a *EnapterAPIv1Adapter) QueryTimeseries(
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

func (a *EnapterAPIv1Adapter) ExecuteCommand(
	ctx context.Context, req *core.ExecuteCommandRequest,
) (*core.ExecuteCommandResponse, error) {
	cmdResp, err := a.commandsAPIClient.Execute(ctx, commandsapi.ExecuteParams{
		User: req.User,
		Request: commandsapi.CommandRequest{
			CommandName: req.CommandName,
			CommandArgs: req.CommandArgs,
			DeviceID:    req.DeviceID,
			HardwareID:  req.HardwareID,
		},
	})
	if err != nil {
		return nil, err
	}
	return &core.ExecuteCommandResponse{
		State:   cmdResp.State,
		Payload: cmdResp.Payload,
	}, nil
}

func (a *EnapterAPIv1Adapter) GetDeviceManifest(
	ctx context.Context, req *core.GetDeviceManifestRequest,
) (*core.GetDeviceManifestResponse, error) {
	device, err := a.assetsAPIClient.DeviceByID(ctx, assetsapi.DeviceByIDParams{
		User:     req.User,
		DeviceID: req.DeviceID,
		Expand: assetsapi.ExpandDeviceParams{
			Manifest: true,
		},
	})
	if err != nil {
		return nil, err
	}
	return &core.GetDeviceManifestResponse{
		Manifest: device.Manifest,
	}, nil
}

func (a *EnapterAPIv1Adapter) convertMultiError(
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
