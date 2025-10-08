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

	"github.com/Enapter/grafana-plugins/pkg/core"
	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi"
)

var _ backend.QueryDataHandler = (*QueryData)(nil)

type QueryData struct {
	logger     hclog.Logger
	enapterAPI core.EnapterAPIPort
}

func NewQueryData(logger hclog.Logger, enapterAPI core.EnapterAPIPort) *QueryData {
	return &QueryData{
		logger:     logger.Named("query_handler"),
		enapterAPI: enapterAPI,
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

	resp, err := h.enapterAPI.ExecuteCommand(ctx, &core.ExecuteCommandRequest{
		User:        user,
		CommandName: props.Payload.CommandName,
		CommandArgs: props.Payload.CommandArgs,
		DeviceID:    props.Payload.DeviceID,
		HardwareID:  props.Payload.HardwareID,
	})
	if err != nil {
		return nil, fmt.Errorf("execute command: %w", err)
	}

	frame, err := h.commandResponseToDataFrame(resp)
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

	resp, err := h.enapterAPI.GetDeviceManifest(ctx, &core.GetDeviceManifestRequest{
		User:     user,
		DeviceID: props.Payload.DeviceID,
	})
	if err != nil {
		return nil, fmt.Errorf("get device manifest: %w", err)
	}

	return data.Frames{
		&data.Frame{Fields: data.Fields{
			data.NewField("manifest", nil, []json.RawMessage{resp.Manifest}),
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

	preparedQuery, err := h.prepareQuery(props.Text, query.Interval, query.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("prepare query text: %w", err)
	}

	resp, err := h.enapterAPI.QueryTimeseries(ctx, &core.QueryTimeseriesRequest{
		User:  user,
		Query: preparedQuery.text,
	})
	if err != nil {
		if errors.Is(err, core.ErrTimeseriesEmpty) {
			return nil, nil
		}
		return nil, fmt.Errorf("query timeseries: %w", err)
	}
	timeseries := resp.Timeseries
	if offset := preparedQuery.offset; offset != 0 {
		timeseries = timeseries.ShiftTime(preparedQuery.offset)
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
	if errors.Is(err, ErrInvalidOffset) {
		return err
	}

	if e := (&yaml.TypeError{}); errors.As(err, &e) {
		return ErrInvalidYAML
	}

	var multiErr *enapterapi.MultiError

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

type preparedQuery struct {
	text   string
	offset time.Duration
}

func (h *QueryData) prepareQuery(
	text string, interval time.Duration, timeRange backend.TimeRange,
) (*preparedQuery, error) {
	dec := yaml.NewDecoder(strings.NewReader(text))

	var obj map[string]interface{}
	if err := dec.Decode(&obj); err != nil {
		return nil, fmt.Errorf("decode YAML: %w", err)
	}

	from := timeRange.From
	to := timeRange.To
	var offset time.Duration
	if offsetInterface, ok := obj["@offset"]; ok {
		offsetString, ok := offsetInterface.(string)
		if !ok {
			return nil, fmt.Errorf("%w: unexpected type: want %T, have %T",
				ErrInvalidOffset, offsetString, offsetInterface)
		}
		var err error
		offset, err = time.ParseDuration(offsetString)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidOffset, err)
		}
		from = from.Add(-offset)
		to = to.Add(-offset)
		delete(obj, "@offset")
	}

	obj["from"] = from.Format(time.RFC3339Nano)
	obj["to"] = to.Format(time.RFC3339Nano)

	if _, ok := obj["granularity"]; !ok {
		obj["granularity"] = h.DefaultGranularity(interval).String()
	}

	if _, ok := obj["aggregation"]; !ok {
		obj["aggregation"] = "auto"
	}

	out, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("encode JSON: %w", err)
	}

	return &preparedQuery{
		text:   string(out),
		offset: offset,
	}, nil
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
	resp *core.ExecuteCommandResponse,
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

func (h *QueryData) timeseriesToDataFrame(
	timeseries *core.Timeseries,
) (*data.Frame, error) {
	frameFields := make([]*data.Field, len(timeseries.DataFields)+1)

	frameFields[0] = data.NewField("time", nil, timeseries.TimeField)

	for i, dataField := range timeseries.DataFields {
		var frameField *data.Field

		switch dataField.Type {
		case core.TimeseriesDataTypeFloat:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*float64, len(dataField.Values)))
		case core.TimeseriesDataTypeInteger:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*int64, len(dataField.Values)))
		case core.TimeseriesDataTypeString:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*string, len(dataField.Values)))
		case core.TimeseriesDataTypeBoolean:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*bool, len(dataField.Values)))
		default:
			return nil, fmt.Errorf("%w: %T",
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
