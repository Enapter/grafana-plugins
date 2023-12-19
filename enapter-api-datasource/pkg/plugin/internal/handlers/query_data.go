package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/hashicorp/go-hclog"
	"gopkg.in/yaml.v3"

	"github.com/Enapter/grafana-plugins/pkg/assetsapi"
	"github.com/Enapter/grafana-plugins/pkg/commandsapi"
	"github.com/Enapter/grafana-plugins/pkg/httperr"
	"github.com/Enapter/grafana-plugins/pkg/telemetryapi"
)

var _ backend.QueryDataHandler = (*QueryData)(nil)

type QueryData struct {
	logger             hclog.Logger
	telemetryAPIClient telemetryapi.Client
	commandsAPIClient  commandsapi.Client
	assetsAPIClient    assetsapi.Client
}

func NewQueryData(
	logger hclog.Logger, telemetryAPIClient telemetryapi.Client,
	commandsAPIClient commandsapi.Client, assetsAPIClient assetsapi.Client,
) *QueryData {
	return &QueryData{
		logger:             logger.Named("query_handler"),
		telemetryAPIClient: telemetryAPIClient,
		commandsAPIClient:  commandsAPIClient,
		assetsAPIClient:    assetsAPIClient,
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

func (h *QueryData) handleQuery(
	ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery,
) (data.Frames, error) {
	var handler func(
		context.Context, backend.PluginContext, backend.DataQuery,
	) (data.Frames, error)

	switch query.QueryType {
	case "command":
		handler = h.handleCommandQuery
	case "manifest":
		handler = h.handleManifestQuery
	case "", "telemetry":
		handler = h.handleTelemetryQuery
	default:
		return nil, errUnexpectedQueryType
	}

	frames, err := handler(ctx, pCtx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", query.QueryType, err)
	}

	return frames, nil
}

func (h *QueryData) handleCommandQuery(
	ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery,
) (data.Frames, error) {
	user := ""
	if pCtx.User != nil {
		user = pCtx.User.Email
	}

	//nolint:tagliatelle // js
	var props struct {
		Payload struct {
			CommandName string                 `json:"commandName"`
			CommandArgs map[string]interface{} `json:"commandArgs"`
			DeviceID    string                 `json:"deviceId"`
			HardwareID  string                 `json:"hardwareId"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(query.JSON, &props); err != nil {
		return nil, fmt.Errorf("parse query properties: %w", err)
	}

	cmdResp, err := h.commandsAPIClient.Execute(ctx, commandsapi.ExecuteParams{
		User: user,
		Request: commandsapi.CommandRequest{
			CommandName: props.Payload.CommandName,
			CommandArgs: props.Payload.CommandArgs,
			DeviceID:    props.Payload.DeviceID,
			HardwareID:  props.Payload.HardwareID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("commands api: %w", err)
	}

	frame, err := h.commandResponseToDataFrame(cmdResp)
	if err != nil {
		return nil, fmt.Errorf("convert cmd resp to data frame: %w", err)
	}

	return data.Frames{frame}, nil
}

func (h *QueryData) handleManifestQuery(
	ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery,
) (data.Frames, error) {
	user := ""
	if pCtx.User != nil {
		user = pCtx.User.Email
	}

	//nolint:tagliatelle // js
	var props struct {
		Payload struct {
			DeviceID string `json:"deviceId"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(query.JSON, &props); err != nil {
		return nil, fmt.Errorf("parse query properties: %w", err)
	}

	device, err := h.assetsAPIClient.DeviceByID(
		ctx, assetsapi.DeviceByIDParams{
			User:     user,
			DeviceID: props.Payload.DeviceID,
			Expand: assetsapi.ExpandDeviceParams{
				Manifest: true,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("assets api: %w", err)
	}

	return data.Frames{
		&data.Frame{Fields: data.Fields{
			data.NewField("manifest", nil, []json.RawMessage{device.Manifest}),
		}},
	}, nil
}

func (h *QueryData) handleTelemetryQuery(
	ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery,
) (data.Frames, error) {
	user := ""
	if pCtx.User != nil {
		user = pCtx.User.Email
	}

	var props struct {
		Hide bool   `json:"hide"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(query.JSON, &props); err != nil {
		return nil, fmt.Errorf("parse query properties: %w", err)
	}

	if props.Hide || len(props.Text) == 0 {
		return nil, nil
	}

	queryText, err := h.prepareQueryText(props.Text, query.Interval, query.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("prepare query text: %w", err)
	}

	timeseries, err := h.telemetryAPIClient.Timeseries(ctx, telemetryapi.TimeseriesParams{
		User:  user,
		Query: queryText,
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

	h.makeLabelsUnique(frame)

	return data.Frames{frame}, nil
}

func (h *QueryData) userFacingError(err error) error {
	if errors.Is(err, errUnsupportedTimeseriesDataType) {
		return ErrMetricDataTypeIsNotSupported
	}

	if e := (&yaml.TypeError{}); errors.As(err, &e) {
		return ErrInvalidYAML
	}

	var multiErr *httperr.MultiError

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

func (h *QueryData) prepareQueryText(
	text string, interval time.Duration, timeRange backend.TimeRange,
) (string, error) {
	dec := yaml.NewDecoder(strings.NewReader(text))

	var obj map[string]interface{}
	if err := dec.Decode(&obj); err != nil {
		return "", fmt.Errorf("decode YAML: %w", err)
	}

	obj["from"] = timeRange.From.Format(time.RFC3339Nano)
	obj["to"] = timeRange.To.Format(time.RFC3339Nano)

	if _, ok := obj["granularity"]; !ok {
		obj["granularity"] = h.DefaultGranularity(interval).String()
	}

	if _, ok := obj["aggregation"]; !ok {
		obj["aggregation"] = "auto"
	}

	out, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("encode JSON: %w", err)
	}

	return string(out), nil
}

func (h *QueryData) DefaultGranularity(interval time.Duration) time.Duration {
	const minInterval = time.Second
	if interval <= minInterval {
		return minInterval
	}

	granularities := [...]time.Duration{
		time.Second,
		2 * time.Second,
		5 * time.Second,
		time.Minute,
		2 * time.Minute,
		5 * time.Minute,
		10 * time.Minute,
		20 * time.Minute,
		30 * time.Minute,
		time.Hour,
		2 * time.Hour,
		6 * time.Hour,
		12 * time.Hour,
		24 * time.Hour,
	}
	if i := sort.Search(len(granularities), func(i int) bool {
		return interval <= granularities[i]
	}); i < len(granularities) {
		return granularities[i]
	}

	return granularities[len(granularities)-1]
}

func (h *QueryData) commandResponseToDataFrame(
	resp commandsapi.CommandResponse,
) (*data.Frame, error) {
	payloadBytes, err := json.Marshal(resp.Payload)
	if err != nil {
		return nil, err
	}

	return &data.Frame{Fields: data.Fields{
		data.NewField("state", nil, []string{resp.State}),
		data.NewField("payload", nil, []json.RawMessage{payloadBytes}),
	}}, nil
}

func (h *QueryData) timeseriesToDataFrame(timeseries *telemetryapi.Timeseries) (*data.Frame, error) {
	frameFields := make([]*data.Field, len(timeseries.DataFields)+1)

	frameFields[0] = data.NewField("time", nil, timeseries.TimeField)

	for i, dataField := range timeseries.DataFields {
		var frameField *data.Field

		switch dataField.Type {
		case telemetryapi.TimeseriesDataTypeFloat:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*float64, len(dataField.Values)))
		case telemetryapi.TimeseriesDataTypeInteger:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*int64, len(dataField.Values)))
		case telemetryapi.TimeseriesDataTypeString:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*string, len(dataField.Values)))
		case telemetryapi.TimeseriesDataTypeBoolean:
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

func (h *QueryData) makeLabelsUnique(frame *data.Frame) {
	if len(frame.Fields) == 0 {
		return
	}

	const oneForTimeField = 1
	dataFields := frame.Fields[oneForTimeField:]

	const metricLabelName = "telemetry"
	var defaultName string

	type kvpair struct {
		k string
		v string
	}
	counter := make(map[kvpair]int)

	for _, field := range dataFields {
		for k, v := range field.Labels {
			if k == metricLabelName {
				defaultName = v
			}
			counter[kvpair{k, v}]++
		}
	}

	for kv, n := range counter {
		if n == len(dataFields) {
			for _, field := range dataFields {
				delete(field.Labels, kv.k)
			}
		}
	}

	for _, field := range dataFields {
		if len(field.Labels) == 0 && defaultName != "" {
			field.Labels = data.Labels{
				metricLabelName: defaultName,
			}
		}
	}
}
