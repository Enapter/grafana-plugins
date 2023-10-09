package telemetryapi

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Enapter/grafana-plugins/pkg/httperr"
)

type Client interface {
	Timeseries(ctx context.Context, p TimeseriesParams) (*Timeseries, error)
	Ready(ctx context.Context) error
	Close()
}

type ClientParams struct {
	HTTPClient *http.Client
	BaseURL    string
	Token      string
}

const DefaultTimeout = 15 * time.Second

func NewClient(p ClientParams) (Client, error) {
	if p.HTTPClient == nil {
		p.HTTPClient = &http.Client{
			Timeout: DefaultTimeout,
		}
	}
	if p.BaseURL == "" {
		p.BaseURL = "https://api.enapter.com/telemetry"
	}

	return &client{
		httpClient: p.HTTPClient,
		baseURL:    p.BaseURL,
		token:      p.Token,
	}, nil
}

type client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

func (c *client) Close() {
	c.httpClient.CloseIdleConnections()
}

func (c *client) Ready(ctx context.Context) error {
	from := time.Now().Add(-time.Hour)
	to := time.Now()
	_, err := c.Timeseries(ctx, TimeseriesParams{
		User: "<not specified>",
		Query: fmt.Sprintf(`{
			"from": %q,
			"to":   %q,
			"telemetry": [{
				"device":    "<does not exist>",
				"attribute": "<does not exist>"
			}],
			"granularity": 	"1m",
			"aggregation":	"auto"
		}`, from.Format(time.RFC3339), to.Format(time.RFC3339)),
	})
	if err == nil {
		return errUnexpectedAbsenceOfError
	}

	var multiErr *httperr.MultiError
	if ok := errors.As(err, &multiErr); !ok || len(multiErr.Errors) != 1 {
		return err
	}

	if multiErr.Errors[0].Code != "unprocessable_entity" {
		return err
	}

	return nil
}

type TimeseriesParams struct {
	User  string
	Query string
}

func (c *client) Timeseries(ctx context.Context, p TimeseriesParams) (_ *Timeseries, retErr error) {
	req, err := c.newTimeseriesRequest(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("new timeseries request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do HTTP request: %w", err)
	}
	defer func() {
		if err := c.drainAndClose(resp.Body); err != nil {
			if retErr == nil {
				retErr = err
			}
		}
	}()

	timeseries, err := c.processTimeseriesResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("process timeseries response: %w", err)
	}

	return timeseries, nil
}

func (c *client) newTimeseriesRequest(ctx context.Context, p TimeseriesParams) (*http.Request, error) {
	urlString := c.baseURL + "/v1/timeseries"

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, urlString, strings.NewReader(p.Query))
	if err != nil {
		return nil, err
	}

	req.Header["Accept"] = []string{"text/csv"}

	if p.User != "" {
		const userField = "X-Enapter-Auth-User"
		req.Header[userField] = []string{p.User}
	}

	const tokenField = "X-Enapter-Auth-Token" //nolint: gosec // false positive
	req.Header[tokenField] = []string{c.token}

	return req, nil
}

func (c *client) processTimeseriesResponse(resp *http.Response) (*Timeseries, error) {
	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusBadRequest, http.StatusForbidden, http.StatusUnprocessableEntity,
		http.StatusTooManyRequests, http.StatusInternalServerError:
		return nil, c.processError(resp)
	default:
		return nil, c.processUnexpectedStatus(resp)
	}

	if s := resp.Header.Get("Content-Length"); s != "" {
		const base = 10
		const bitSize = 64
		n, err := strconv.ParseUint(s, base, bitSize)
		if err != nil {
			return nil, fmt.Errorf("parse content length: %w", err)
		}
		if n == 0 {
			return nil, ErrNoValues
		}
	}

	const wantContentType = "text/csv"
	if have := resp.Header.Get("Content-Type"); have != wantContentType {
		return nil, fmt.Errorf("%w: want %s, have %s",
			errUnexpectedContentType, wantContentType, have)
	}

	const dataTypesField = "X-Enapter-Timeseries-Data-Types"
	if v := resp.Header.Get(dataTypesField); v == "" {
		return nil, fmt.Errorf("%w: %s", errEmptyHeaderField, dataTypesField)
	}

	dataTypeStrings := resp.Header.Values(dataTypesField)

	dataTypes, err := parseTimeseriesDataTypes(dataTypeStrings)
	if err != nil {
		return nil, fmt.Errorf("parse data types: %w", err)
	}

	timeseries, err := c.parseTimeseriesCSV(resp.Body, dataTypes)
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	return timeseries, nil
}

func (c *client) parseTimeseriesCSV(
	reader io.Reader, dataTypes []TimeseriesDataType,
) (*Timeseries, error) {
	csvReader := csv.NewReader(reader)

	const dontCheckNumberOfFields = -1
	csvReader.FieldsPerRecord = dontCheckNumberOfFields
	csvReader.ReuseRecord = true

	timeseries := NewTimeseries(dataTypes)

	for i := 0; ; i++ {
		record, err := csvReader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read record %d: %w", i, err)
		}

		const headerIndex = 0
		if i == headerIndex {
			tagsList, err := c.parseCSVHeader(record, len(dataTypes))
			if err != nil {
				return nil, fmt.Errorf("parse header: %w", err)
			}
			for j, tags := range tagsList {
				timeseries.DataFields[j].Tags = tags
			}
			continue
		}

		timestamp, values, err := c.parseTimeseriesCSVRecord(record, dataTypes)
		if err != nil {
			return nil, fmt.Errorf("parse record %d: %w", i, err)
		}

		timeseries.TimeField = append(timeseries.TimeField, timestamp)
		for j, value := range values {
			dataField := timeseries.DataFields[j]
			dataField.Values = append(dataField.Values, value)
		}
	}

	if timeseries.Len() == 0 {
		return nil, ErrNoValues
	}

	return timeseries, nil
}

func (c *client) parseCSVHeader(record []string, numDataFields int) ([]TimeseriesTags, error) {
	if want := (numDataFields + 1); len(record) != want {
		return nil, fmt.Errorf("%w: want %d, have %d",
			errUnexpectedNumberOfFields, want, len(record))
	}

	const ts = "ts"
	if record[0] != ts {
		return nil, fmt.Errorf("%w: want %s, have %s",
			errUnexpectedFieldName, ts, record[0])
	}

	tagsList := make([]TimeseriesTags, numDataFields)

	for i := 0; i < numDataFields; i++ {
		tags, err := parseTimeseriesTags(record[i+1])
		if err != nil {
			return nil, fmt.Errorf("tags: %w", err)
		}

		tagsList[i] = tags
	}

	return tagsList, nil
}

func (c *client) parseTimeseriesCSVRecord(
	record []string, dataTypes []TimeseriesDataType,
) (time.Time, []interface{}, error) {
	if want := len(dataTypes) + 1; len(record) != want {
		return time.Time{}, nil, fmt.Errorf("%w: want %d, have %d",
			errUnexpectedNumberOfFields, want, len(record))
	}

	const base = 10
	const bitSize = 64
	timestamp, err := strconv.ParseInt(record[0], base, bitSize)
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("timestamp: %w", err)
	}

	values := make([]interface{}, len(dataTypes))

	for i := 0; i < len(dataTypes); i++ {
		field := record[i+1]
		value, err := dataTypes[i].Parse(field)
		if err != nil {
			return time.Time{}, nil, fmt.Errorf("field %d: %w", i, err)
		}
		values[i] = value
	}

	return time.Unix(timestamp, 0), values, nil
}

func (c *client) drainAndClose(rc io.ReadCloser) error {
	_, _ = io.Copy(io.Discard, rc)
	return rc.Close()
}

func (c *client) processError(resp *http.Response) error {
	multiErr, err := httperr.ParseMultiError(resp.Body)
	if err != nil {
		return fmt.Errorf("multi-error: <not available>: %w", err)
	}

	return multiErr
}

func (c *client) processUnexpectedStatus(resp *http.Response) error {
	dump, err := dumpBody(resp.Body)
	if err != nil {
		//nolint:errorlint // two errors
		return fmt.Errorf("%w: %s: body dump: <not available>: %v",
			errUnexpectedStatus, resp.Status, err)
	}

	return fmt.Errorf("%w: %s: body dump: %s",
		errUnexpectedStatus, resp.Status, dump)
}
