package handlers

import (
	"context"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/hashicorp/go-hclog"

	"github.com/Enapter/grafana-plugins/pkg/core"
)

var _ backend.CheckHealthHandler = (*CheckHealth)(nil)

type CheckHealth struct {
	logger     hclog.Logger
	enapterAPI core.EnapterAPIPort
}

func NewCheckHealth(logger hclog.Logger, enapterAPI core.EnapterAPIPort) *CheckHealth {
	return &CheckHealth{
		logger:     logger.Named("check_health_handler"),
		enapterAPI: enapterAPI,
	}
}

func (h *CheckHealth) CheckHealth(
	ctx context.Context, _ *backend.CheckHealthRequest,
) (*backend.CheckHealthResult, error) {
	if err := h.enapterAPI.Ready(ctx); err != nil {
		h.logger.Error("Enapter API is not ready", "error", err.Error())

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
