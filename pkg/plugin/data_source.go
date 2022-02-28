package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/hashicorp/go-hclog"

	"github.com/Enapter/grafana-plugins/telemetry-datasource/pkg/plugin/internal/queryhandler"
	"github.com/Enapter/grafana-plugins/telemetry-datasource/pkg/telemetryapi"
)

var (
	_ backend.QueryDataHandler      = (*dataSource)(nil)
	_ backend.CheckHealthHandler    = (*dataSource)(nil)
	_ instancemgmt.InstanceDisposer = (*dataSource)(nil)
)

type dataSource struct {
	logger             hclog.Logger
	telemetryAPIClient telemetryapi.Client
	queryHandler       *queryhandler.QueryHandler
}

func newDataSource(logger hclog.Logger, settings backend.DataSourceInstanceSettings,
) (inst instancemgmt.Instance, retErr error) {
	logger = logger.Named(fmt.Sprintf("data_source[%q]", settings.Name))

	defer func() {
		if retErr != nil {
			logger.Error("failed to create data source",
				"error", retErr.Error())
		}
	}()

	var jsonData map[string]string
	if err := json.Unmarshal(settings.JSONData, &jsonData); err != nil {
		return nil, fmt.Errorf("JSON data: %w", err)
	}

	telemetryAPIClient, err := telemetryapi.NewClient(telemetryapi.ClientParams{
		Logger:  logger,
		BaseURL: jsonData["telemetryAPIBaseURL"],
	})
	if err != nil {
		return nil, fmt.Errorf("new telemetry API client: %w", err)
	}

	queryHandler := queryhandler.New(telemetryAPIClient)

	logger.Info("created new data source")

	return &dataSource{
		logger:             logger,
		telemetryAPIClient: telemetryAPIClient,
		queryHandler:       queryHandler,
	}, nil
}

func (d *dataSource) Dispose() {
	d.telemetryAPIClient.Close()

	d.logger.Info("disposed data source")
}

func (d *dataSource) QueryData(
	ctx context.Context, req *backend.QueryDataRequest,
) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		frames, err := d.queryHandler.HandleQuery(ctx, req.PluginContext, q)
		if err != nil {
			d.logger.Warn("failed to handle query",
				"ref_id", q.RefID,
				"error", err)

			err = d.userFacingError(err)
		}

		resp.Responses[q.RefID] = backend.DataResponse{
			Frames: frames,
			Error:  err,
		}
	}

	return resp, nil
}

func (d *dataSource) userFacingError(err error) error {
	var multiErr *telemetryapi.MultiError

	if ok := errors.As(err, &multiErr); !ok {
		return errSomethingWentWrong
	}

	switch len(multiErr.Errors) {
	case 0: // should never happen
		d.logger.Error("multi error does not contains errors")
		return errSomethingWentWrong
	case 1:
		// noop
	default:
		d.logger.Warn("multi error contains multiple errors, " +
			"but this is not supported yet; will return only the first error")
	}

	if msg := multiErr.Errors[0].Message; len(msg) > 0 {
		//nolint: goerr113 // user-facing
		return errors.New(msg)
	}

	return errSomethingWentWrong
}

func (d *dataSource) CheckHealth(
	ctx context.Context, req *backend.CheckHealthRequest,
) (*backend.CheckHealthResult, error) {
	if err := d.telemetryAPIClient.Ready(ctx); err != nil {
		d.logger.Error("telemetry API client is not ready",
			"error", err.Error())

		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "ok",
	}, nil
}
