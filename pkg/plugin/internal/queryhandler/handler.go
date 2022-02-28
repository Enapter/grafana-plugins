package queryhandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

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

	const fieldTimestampName = "timestamp"
	const fieldValueName = "value"

	fieldTimestamp := data.NewField(fieldTimestampName, nil,
		make([]time.Time, len(timeseries.Values)))

	var fieldValues *data.Field
	switch timeseries.DataType {
	case telemetryapi.TimeseriesDataTypeFloat64:
		fieldValues = data.NewField(fieldValueName, nil,
			make([]float64, len(timeseries.Values)))
	case telemetryapi.TimeseriesDataTypeInt64:
		fieldValues = data.NewField(fieldValueName, nil,
			make([]int64, len(timeseries.Values)))
	case telemetryapi.TimeseriesDataTypeString:
		fieldValues = data.NewField(fieldValueName, nil,
			make([]string, len(timeseries.Values)))
	case telemetryapi.TimeseriesDataTypeBool:
		fieldValues = data.NewField(fieldValueName, nil,
			make([]bool, len(timeseries.Values)))
	default:
		return nil, fmt.Errorf("%w: %s",
			errUnsupportedTimeseriesDataType, timeseries.DataType)
	}

	for i, v := range timeseries.Values {
		fieldTimestamp.Set(i, v.Timestamp)
		fieldValues.Set(i, v.Value)
	}
	frame := data.NewFrame("response", fieldTimestamp, fieldValues)

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
