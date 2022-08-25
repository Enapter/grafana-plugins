package handlers

import (
	"context"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/hashicorp/go-hclog"

	"github.com/Enapter/telemetry-grafana-datasource-plugin/pkg/telemetryapi"
)

var _ backend.CheckHealthHandler = (*CheckHealth)(nil)

type CheckHealth struct {
	logger hclog.Logger
	client telemetryapi.Client
}

func NewCheckHealth(logger hclog.Logger, client telemetryapi.Client) *CheckHealth {
	return &CheckHealth{
		logger: logger.Named("check_health_handler"),
		client: client,
	}
}

func (h *CheckHealth) CheckHealth(
	ctx context.Context, req *backend.CheckHealthRequest,
) (*backend.CheckHealthResult, error) {
	if err := h.client.Ready(ctx); err != nil {
		h.logger.Error("telemetry API client is not ready",
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
