package telemetryapi

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
)

type Client interface {
	Timeseries(ctx context.Context, p TimeseriesParams) (*Timeseries, error)
	Ready(ctx context.Context) error
	Close()
}

type ClientParams struct {
	Logger     hclog.Logger
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
		logger:     p.Logger.Named("telemetry_api_client"),
		httpClient: p.HTTPClient,
		baseURL:    p.BaseURL,
		token:      p.Token,
	}, nil
}

type client struct {
	logger     hclog.Logger
	httpClient *http.Client
	baseURL    string
	token      string
}

func (c *client) Close() {
	c.httpClient.CloseIdleConnections()
}

func (c *client) Ready(ctx context.Context) error {
	req, err := c.newReadyRequest(ctx)
	if err != nil {
		return fmt.Errorf("new ready request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do HTTP request: %w", err)
	}
	defer c.closeRespBody(resp.Body)

	return c.processReadyResponse(resp)
}

func (c *client) newReadyRequest(ctx context.Context) (*http.Request, error) {
	urlString := c.baseURL + "/_/ready"
	return http.NewRequestWithContext(ctx, http.MethodGet, urlString, nil)
}

func (c *client) processReadyResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return c.processUnexpectedStatus(resp)
	}

	return nil
}

type TimeseriesParams struct {
	User  string
	Query string
	From  time.Time
	To    time.Time
}

func (c *client) Timeseries(ctx context.Context, p TimeseriesParams) (*Timeseries, error) {
	req, err := c.newTimeseriesRequest(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("new timeseries request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do HTTP request: %w", err)
	}
	defer c.closeRespBody(resp.Body)

	timeseries, err := c.processTimeseriesResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("process timeseries response: %w", err)
	}

	return timeseries, nil
}

func (c *client) newTimeseriesRequest(ctx context.Context, p TimeseriesParams) (*http.Request, error) {
	if p.User == "" {
		return nil, ErrEmptyUser
	}

	q := make(url.Values)

	q.Set("from", p.From.UTC().Format(time.RFC3339))
	q.Set("to", p.To.UTC().Format(time.RFC3339))

	urlString := c.baseURL + "/v1/timeseries?" + q.Encode()

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, urlString, strings.NewReader(p.Query))
	if err != nil {
		return nil, err
	}

	const userField = "X-Enapter-Auth-User"
	req.Header[userField] = []string{p.User}

	const tokenField = "X-Enapter-Auth-Token"
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
	dataTypeString := resp.Header.Get(dataTypesField)
	if dataTypeString == "" {
		return nil, fmt.Errorf("%w: %s", errEmptyHeaderField, dataTypesField)
	}

	dataType, err := parseTimeseriesDataType(dataTypeString)
	if err != nil {
		return nil, fmt.Errorf("parse data type: %w", err)
	}

	timeseries, err := c.parseTimeseriesCSV(resp.Body, dataType)
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	return timeseries, nil
}

func (c *client) parseTimeseriesCSV(reader io.Reader, dataType TimeseriesDataType) (*Timeseries, error) {
	csvReader := csv.NewReader(reader)

	const dontCheckNumberOfFields = -1
	csvReader.FieldsPerRecord = dontCheckNumberOfFields
	csvReader.ReuseRecord = true

	const preallocValues = 64
	timeseries := &Timeseries{
		Values:   make([]*TimeseriesValue, 0, preallocValues),
		DataType: dataType,
	}

	for i := 0; ; i++ {
		record, err := csvReader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read record %d: %w", i, err)
		}

		const skipCSVHeader = true
		if i == 0 && skipCSVHeader {
			if len(record) < 2 || record[0] != "ts" {
				return nil, fmt.Errorf("parse record %d: %w",
					i, errUnexpectedCSVHeader)
			}
			continue
		}

		value, err := c.parseTimeseriesCSVRecord(record, dataType)
		if err != nil {
			return nil, fmt.Errorf("parse record %d: %w", i, err)
		}
		timeseries.Values = append(timeseries.Values, value)
	}

	if len(timeseries.Values) == 0 {
		return nil, ErrNoValues
	}

	return timeseries, nil
}

func (c *client) parseTimeseriesCSVRecord(
	record []string, dataType TimeseriesDataType,
) (*TimeseriesValue, error) {
	const (
		iTimestamp = iota
		iValue
		nFields
	)

	if len(record) != nFields {
		return nil, fmt.Errorf("%w: want %d, have %d",
			errUnexpectedNumberOfFields, nFields, len(record))
	}

	const base = 10
	const bitSize = 64
	timestamp, err := strconv.ParseInt(record[iTimestamp], base, bitSize)
	if err != nil {
		return nil, fmt.Errorf("timestamp: %w", err)
	}

	data, err := dataType.Parse(record[iValue])
	if err != nil {
		return nil, fmt.Errorf("data: %w", err)
	}

	return &TimeseriesValue{
		Timestamp: time.Unix(timestamp, 0),
		Value:     data,
	}, nil
}

func (c *client) closeRespBody(body io.ReadCloser) {
	if _, err := io.Copy(io.Discard, body); err != nil {
		c.logger.Warn("failed to drain HTTP response body",
			"error", err.Error())
	}

	if err := body.Close(); err != nil {
		c.logger.Warn("failed to close HTTP response body",
			"error", err.Error())
	}
}

func (c *client) processError(resp *http.Response) error {
	multiErr, err := parseMultiError(resp.Body)
	if err != nil {
		return fmt.Errorf("multi-error: <not available>: %w", err)
	}

	return multiErr
}

func (c *client) processUnexpectedStatus(resp *http.Response) error {
	dump, err := dumpBody(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: %s: body dump: <not available>: %v",
			errUnexpectedStatus, resp.Status, err)
	}

	return fmt.Errorf("%w: %s: body dump: %s",
		errUnexpectedStatus, resp.Status, dump)
}
