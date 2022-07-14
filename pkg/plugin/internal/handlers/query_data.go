package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/hashicorp/go-hclog"

	"github.com/Enapter/grafana-plugins/telemetry-datasource/pkg/telemetryapi"
)

var _ backend.QueryDataHandler = (*QueryData)(nil)

type QueryData struct {
	logger             hclog.Logger
	telemetryAPIClient telemetryapi.Client
}

func NewQueryData(logger hclog.Logger, telemetryAPIClient telemetryapi.Client) *QueryData {
	return &QueryData{
		logger:             logger.Named("query_handler"),
		telemetryAPIClient: telemetryAPIClient,
	}
}

func (h *QueryData) QueryData(
	ctx context.Context, req *backend.QueryDataRequest,
) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		frames, err := h.handleQuery(ctx, req.PluginContext, q)
		if err != nil {
			h.logger.Warn("failed to handle query",
				"ref_id", q.RefID,
				"error", err)

			err = h.userFacingError(err)
		}

		resp.Responses[q.RefID] = backend.DataResponse{
			Frames: frames,
			Error:  err,
		}
	}

	return resp, nil
}

func (h *QueryData) userFacingError(err error) error {
	if errors.Is(err, errUnsupportedTimeseriesDataType) {
		return ErrMetricDataTypeIsNotSupported
	}

	var multiErr *telemetryapi.MultiError

	if ok := errors.As(err, &multiErr); !ok {
		return ErrSomethingWentWrong
	}

	switch len(multiErr.Errors) {
	case 0: // should never happen
		h.logger.Error("multi error does not contains errors")
		return ErrSomethingWentWrong
	case 1:
		// noop
	default:
		h.logger.Warn("multi error contains multiple errors, " +
			"but this is not supported yet; will return only the first error")
	}

	if msg := multiErr.Errors[0].Message; len(msg) > 0 {
		//nolint: goerr113 // user-facing
		return errors.New(msg)
	}

	return ErrSomethingWentWrong
}

func (h *QueryData) handleQuery(
	ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery,
) (data.Frames, error) {
	user := ""
	if pCtx.User != nil {
		user = pCtx.User.Email
	}

	var props queryProperties
	if err := json.Unmarshal(query.JSON, &props); err != nil {
		return nil, fmt.Errorf("parse query properties: %w", err)
	}

	if props.Hide || len(props.Text) == 0 {
		return nil, nil
	}

	queryText, err := h.parseQueryText(props.Text)
	if err != nil {
		return nil, fmt.Errorf("parse query text: %w", err)
	}

	timeseries, err := h.telemetryAPIClient.Timeseries(ctx, telemetryapi.TimeseriesParams{
		User:  user,
		Query: queryText,
		From:  query.TimeRange.From,
		To:    query.TimeRange.To,
	})
	if err != nil {
		if errors.Is(err, telemetryapi.ErrNoValues) {
			return nil, nil
		}
		return nil, fmt.Errorf("request timeseries: %w", err)
	}

	frame, err := h.timeseriesToDataFrame(timeseries)
	if err != nil {
		return nil, fmt.Errorf("convert timeseries to data frame: %w", err)
	}

	return data.Frames{frame}, nil
}

func (h *QueryData) parseQueryText(text string) (string, error) {
	parsed, err := yamlToJSON(text)
	if err != nil {
		return "", fmt.Errorf("convert YAML to JSON: %w", err)
	}

	return parsed, nil
}

func (h *QueryData) timeseriesToDataFrame(timeseries *telemetryapi.Timeseries) (*data.Frame, error) {
	frameFields := make([]*data.Field, len(timeseries.DataFields)+1)

	frameFields[0] = data.NewField("time", nil, timeseries.TimeField)

	for i, dataField := range timeseries.DataFields {
		var frameField *data.Field

		switch dataField.Type {
		case telemetryapi.TimeseriesDataTypeFloat64:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*float64, len(dataField.Values)))
		case telemetryapi.TimeseriesDataTypeInt64:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*int64, len(dataField.Values)))
		case telemetryapi.TimeseriesDataTypeString:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*string, len(dataField.Values)))
		case telemetryapi.TimeseriesDataTypeBool:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*bool, len(dataField.Values)))
		default:
			return nil, fmt.Errorf("%w: %s",
				errUnsupportedTimeseriesDataType, dataField.Type)
		}

		frameFields[i+1] = frameField
	}

	for row := 0; row < timeseries.Len(); row++ {
		for col := 0; col < len(timeseries.DataFields); col++ {
			frameField := frameFields[col+1]
			dataField := timeseries.DataFields[col]
			value := dataField.Values[row]
			frameField.Set(row, value)
		}
	}

	return data.NewFrame("", frameFields...), nil
}
