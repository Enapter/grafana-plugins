package core

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
)

var (
	_ backend.CheckHealthHandler = (*DataSource)(nil)
	_ backend.QueryDataHandler   = (*DataSource)(nil)
)

type DataSource struct {
	logger     hclog.Logger
	enapterAPI EnapterAPIPort
}

type DataSourceParams struct {
	Logger     hclog.Logger
	EnapterAPI EnapterAPIPort
}

func NewDataSource(p DataSourceParams) *DataSource {
	return &DataSource{
		logger:     p.Logger,
		enapterAPI: p.EnapterAPI,
	}
}

func (d *DataSource) CheckHealth(
	ctx context.Context, _ *backend.CheckHealthRequest,
) (*backend.CheckHealthResult, error) {
	if err := d.enapterAPI.Ready(ctx); err != nil {
		d.logger.Error("Enapter API is not ready", "error", err.Error())
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

func (d *DataSource) QueryData(
	ctx context.Context, req *backend.QueryDataRequest,
) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		frames, err := d.handleQuery(ctx, req.PluginContext, q)
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

func (d *DataSource) handleQuery(
	ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery,
) (data.Frames, error) {
	var handler func(
		context.Context, backend.PluginContext, backend.DataQuery,
	) (data.Frames, error)

	queryType := "telemetry"
	if t := query.QueryType; t != "" {
		queryType = t
	}

	switch queryType {
	case "command":
		handler = d.handleCommandQuery
	case "manifest":
		handler = d.handleManifestQuery
	case "telemetry":
		handler = d.handleTelemetryQuery
	default:
		return nil, errUnexpectedQueryType
	}

	frames, err := handler(ctx, pCtx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", queryType, err)
	}

	return frames, nil
}

func (d *DataSource) handleCommandQuery(
	ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery,
) (data.Frames, error) {
	user := ""
	if pCtx.User != nil {
		user = pCtx.User.Email
	}

	//nolint:tagliatelle // js
	var props struct {
		Payload struct {
			CommandName string         `json:"commandName"`
			CommandArgs map[string]any `json:"commandArgs"`
			DeviceID    string         `json:"deviceId"`
			HardwareID  string         `json:"hardwareId"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(query.JSON, &props); err != nil {
		return nil, fmt.Errorf("parse query properties: %w", err)
	}

	resp, err := d.enapterAPI.ExecuteCommand(ctx, &ExecuteCommandRequest{
		User:        user,
		CommandName: props.Payload.CommandName,
		CommandArgs: props.Payload.CommandArgs,
		DeviceID:    props.Payload.DeviceID,
		HardwareID:  props.Payload.HardwareID,
	})
	if err != nil {
		return nil, fmt.Errorf("execute command: %w", err)
	}

	frame, err := d.commandResponseToDataFrame(resp)
	if err != nil {
		return nil, fmt.Errorf("convert cmd resp to data frame: %w", err)
	}

	return data.Frames{frame}, nil
}

func (d *DataSource) handleManifestQuery(
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

	resp, err := d.enapterAPI.GetDeviceManifest(ctx, &GetDeviceManifestRequest{
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

func (d *DataSource) handleTelemetryQuery(
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

	preparedQuery, err := d.prepareQuery(props.Text, query.Interval, query.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("prepare query text: %w", err)
	}

	resp, err := d.enapterAPI.QueryTimeseries(ctx, &QueryTimeseriesRequest{
		User:  user,
		Query: preparedQuery.text,
	})
	if err != nil {
		if errors.Is(err, ErrTimeseriesEmpty) {
			return nil, nil
		}
		return nil, fmt.Errorf("query timeseries: %w", err)
	}
	timeseries := resp.Timeseries
	if offset := preparedQuery.offset; offset != 0 {
		timeseries = timeseries.ShiftTime(preparedQuery.offset)
	}

	frame, err := d.timeseriesToDataFrame(timeseries)
	if err != nil {
		return nil, fmt.Errorf("convert timeseries to data frame: %w", err)
	}

	d.makeLabelsUnique(frame)

	return data.Frames{frame}, nil
}

func (d *DataSource) userFacingError(err error) error {
	if errors.Is(err, errUnsupportedTimeseriesDataType) {
		return ErrMetricDataTypeIsNotSupported
	}
	if errors.Is(err, ErrInvalidOffset) {
		return ErrInvalidOffset
	}

	if e := (&yaml.TypeError{}); errors.As(err, &e) {
		return ErrInvalidYAML
	}

	var apiError EnapterAPIError
	if errors.As(err, &apiError) && len(apiError.Message) > 0 {
		//nolint: goerr113 // user-facing
		return errors.New(apiError.Message)
	}

	return ErrSomethingWentWrong
}

type preparedQuery struct {
	text   string
	offset time.Duration
}

func (d *DataSource) prepareQuery(
	text string, interval time.Duration, timeRange backend.TimeRange,
) (*preparedQuery, error) {
	dec := yaml.NewDecoder(strings.NewReader(text))

	var obj map[string]any
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
		obj["granularity"] = d.DefaultGranularity(interval).String()
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

func (d *DataSource) DefaultGranularity(interval time.Duration) time.Duration {
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

func (d *DataSource) commandResponseToDataFrame(
	resp *ExecuteCommandResponse,
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

func (d *DataSource) timeseriesToDataFrame(
	timeseries *Timeseries,
) (*data.Frame, error) {
	frameFields := make([]*data.Field, len(timeseries.DataFields)+1)

	frameFields[0] = data.NewField("time", nil, timeseries.TimeField)

	for i, dataField := range timeseries.DataFields {
		var frameField *data.Field

		switch dataField.Type {
		case TimeseriesDataTypeFloat:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*float64, len(dataField.Values)))
		case TimeseriesDataTypeInteger:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*int64, len(dataField.Values)))
		case TimeseriesDataTypeString:
			frameField = data.NewField(
				"", data.Labels(dataField.Tags),
				make([]*string, len(dataField.Values)))
		case TimeseriesDataTypeBoolean:
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

func (d *DataSource) makeLabelsUnique(frame *data.Frame) {
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
