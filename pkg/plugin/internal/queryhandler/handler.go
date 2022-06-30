package queryhandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	yaml "gopkg.in/yaml.v3"

	"github.com/Enapter/grafana-plugins/telemetry-datasource/pkg/telemetryapi"
)

type QueryHandler struct {
	telemetryAPIClient telemetryapi.Client
}

func New(telemetryAPIClient telemetryapi.Client) *QueryHandler {
	return &QueryHandler{
		telemetryAPIClient: telemetryAPIClient,
	}
}

func (h *QueryHandler) timeseriesToDataFrame(timeseries *telemetryapi.Timeseries) (*data.Frame, error) {
	frameFields := make([]*data.Field, len(timeseries.DataFields)+1)

	frameFields[0] = data.NewField("time", nil, timeseries.TimeField)

	for i, dataField := range timeseries.DataFields {
		var frameField *data.Field

		switch dataField.Type {
		case telemetryapi.TimeseriesDataTypeFloat64:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]float64, len(dataField.Values)))
		case telemetryapi.TimeseriesDataTypeInt64:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]int64, len(dataField.Values)))
		case telemetryapi.TimeseriesDataTypeString:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]string, len(dataField.Values)))
		case telemetryapi.TimeseriesDataTypeBool:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]bool, len(dataField.Values)))
		default:
			return nil, fmt.Errorf("%w: %s",
				ErrUnsupportedTimeseriesDataType, dataField.Type)
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

	return data.NewFrame("response", frameFields...), nil
}

func (h *QueryHandler) HandleQuery(
	ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery,
) (data.Frames, error) {
	if pCtx.User == nil {
		return nil, ErrMissingUserInfo
	}

	queryText, err := h.parseRawQuery(query)
	if err != nil {
		if errors.Is(err, errEmptyQueryText) {
			return nil, nil
		}
		return nil, fmt.Errorf("parse raw query: %w", err)
	}

	timeseries, err := h.telemetryAPIClient.Timeseries(ctx, telemetryapi.TimeseriesParams{
		User:  pCtx.User.Email,
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

func (h *QueryHandler) parseRawQuery(raw backend.DataQuery) (string, error) {
	var query struct {
		TextAsYAML string `json:"text"`
	}
	if err := json.Unmarshal(raw.JSON, &query); err != nil {
		return "", err
	}
	if len(query.TextAsYAML) == 0 {
		return "", errEmptyQueryText
	}

	queryTextAsJSON, err := yamlToJSON(query.TextAsYAML)
	if err != nil {
		return "", fmt.Errorf("convert YAML to JSON: %w", err)
	}

	return queryTextAsJSON, nil
}

func yamlToJSON(in string) (string, error) {
	dec := yaml.NewDecoder(strings.NewReader(in))

	var obj map[string]interface{}
	if err := dec.Decode(&obj); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}

	out, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("encode: %w", err)
	}

	return string(out), nil
}
