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

	"github.com/Enapter/telemetry-grafana-datasource-plugin/pkg/telemetryapi"
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

	h.makeLabelsUnique(resp.Responses)

	return resp, nil
}

const labelTelemetry = "telemetry"

//nolint:gocognit // FIXME
func (h *QueryData) makeLabelsUnique(responses backend.Responses) {
	defaultNames := make(map[string]string, len(responses))
	frames := make(map[string]*data.Frame)
	numFields := 0
	const oneForTimeField = 1

	type kvpair struct {
		k string
		v string
	}

	counter := make(map[kvpair]int)

	for refID, resp := range responses {
		if resp.Error != nil {
			continue
		}

		switch {
		case len(resp.Frames) == 0:
			continue
		case len(resp.Frames) == 1:
			// ok
		case len(resp.Frames) > 1:
			h.logger.Warn("multiple data frames are not supported: " +
				"skip making labels unique")
			return
		default:
			panic("unreachable")
		}

		frame := resp.Frames[0]
		if len(frame.Fields) < oneForTimeField+1 {
			continue
		}

		frames[refID] = frame

		for _, field := range frame.Fields[oneForTimeField:] {
			for k, v := range field.Labels {
				if k == labelTelemetry {
					defaultNames[refID] = v
				}
				counter[kvpair{k, v}]++
			}
			numFields++
		}
	}

	for kv, n := range counter {
		if n == numFields {
			for _, frame := range frames {
				for _, field := range frame.Fields[oneForTimeField:] {
					delete(field.Labels, kv.k)
				}
			}
		}
	}

	for refID, frame := range frames {
		for _, field := range frame.Fields[oneForTimeField:] {
			if len(field.Labels) == 0 {
				field.Name = defaultNames[refID]
			}
		}
	}
}

func (h *QueryData) userFacingError(err error) error {
	if errors.Is(err, errUnsupportedTimeseriesDataType) {
		return ErrMetricDataTypeIsNotSupported
	}

	if e := (&yaml.TypeError{}); errors.As(err, &e) {
		return ErrInvalidYAML
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

	timeseries = h.alertsToFields(timeseries)

	frame, err := h.timeseriesToDataFrame(timeseries)
	if err != nil {
		return nil, fmt.Errorf("convert timeseries to data frame: %w", err)
	}

	return data.Frames{frame}, nil
}

func (h *QueryData) alertsToFields(
	base *telemetryapi.Timeseries,
) *telemetryapi.Timeseries {
	var haveAlerts bool
	for _, dataField := range base.DataFields {
		if isAlertsDataField(dataField) {
			haveAlerts = true
			break
		}
	}
	if !haveAlerts {
		return base
	}

	resultDataFields := make([]*telemetryapi.TimeseriesDataField, 0,
		len(base.DataFields))

	for _, dataField := range base.DataFields {
		if isAlertsDataField(dataField) {
			dataFields := h.splitAlertsDataField(dataField)
			resultDataFields = append(resultDataFields, dataFields...)
		} else {
			resultDataFields = append(resultDataFields, dataField)
		}
	}

	return &telemetryapi.Timeseries{
		TimeField:  base.TimeField,
		DataFields: resultDataFields,
	}
}

const alertsMetricName = "alerts"

func isAlertsDataField(dataField *telemetryapi.TimeseriesDataField) bool {
	return dataField.Tags[labelTelemetry] == alertsMetricName &&
		dataField.Type == telemetryapi.TimeseriesDataTypeStringArray
}

func (h *QueryData) splitAlertsDataField(
	base *telemetryapi.TimeseriesDataField,
) (dataFields []*telemetryapi.TimeseriesDataField) {
	var dataFieldsByAlert map[string]*telemetryapi.TimeseriesDataField

	for i, v := range base.Values {
		//nolint:forcetypeassert // panic is fine
		for _, alert := range v.([]string) {
			dataField, ok := dataFieldsByAlert[alert]
			if !ok {
				if dataFieldsByAlert == nil {
					dataFieldsByAlert = make(map[string]*telemetryapi.TimeseriesDataField)
				}
				dataField = newAlertDataField(base, alert)
				dataFieldsByAlert[alert] = dataField
			}
			*dataField.Values[i].(*bool) = true
		}
	}

	if n := len(dataFieldsByAlert); n > 0 {
		alerts := make([]string, 0, n)
		for alert := range dataFieldsByAlert {
			alerts = append(alerts, alert)
		}
		sort.Strings(alerts)

		dataFields = make([]*telemetryapi.TimeseriesDataField, n)
		for i, alert := range alerts {
			dataFields[i] = dataFieldsByAlert[alert]
		}
	}

	return dataFields
}

func newAlertDataField(
	base *telemetryapi.TimeseriesDataField, name string,
) *telemetryapi.TimeseriesDataField {
	tags := base.Tags.Copy()
	tags[labelTelemetry] = alertsMetricName + "." + name

	values := make([]interface{}, len(base.Values))
	for i := 0; i < len(base.Values); i++ {
		values[i] = new(bool)
	}

	return &telemetryapi.TimeseriesDataField{
		Tags:   tags,
		Type:   telemetryapi.TimeseriesDataTypeBoolean,
		Values: values,
	}
}

func (h *QueryData) prepareQueryText(
	text string, interval time.Duration, timeRange backend.TimeRange,
) (string, error) {
	dec := yaml.NewDecoder(strings.NewReader(text))

	var obj map[string]interface{}
	if err := dec.Decode(&obj); err != nil {
		return "", fmt.Errorf("decode YAML: %w", err)
	}

	obj["from"] = timeRange.From.Format(time.RFC3339)
	obj["to"] = timeRange.To.Format(time.RFC3339)

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
