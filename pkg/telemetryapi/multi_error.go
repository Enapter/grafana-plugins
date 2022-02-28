package telemetryapi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

func parseMultiError(body io.Reader) (*MultiError, error) {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if len(data) == 0 {
		return nil, errEmptyData
	}

	var multiErr MultiError
	if err := json.Unmarshal(data, &multiErr); err != nil {
		return nil, fmt.Errorf("parse data: %w", err)
	}

	if len(multiErr.Errors) == 0 {
		return nil, errEmptyErrorList
	}

	for i, err := range multiErr.Errors {
		if len(err.Code) == 0 {
			multiErr.Errors[i].Code = "<empty>"
		}
	}

	return &multiErr, nil
}

type MultiError struct {
	Errors []Error `json:"errors"`
}

func (m *MultiError) Error() string {
	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}

	return fmt.Sprintf("%d errors: %v", len(m.Errors), m.Errors)
}

type Error struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *Error) Error() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("code=%s", e.Code))

	if len(e.Message) > 0 {
		b.WriteString(fmt.Sprintf(", message=%q", e.Message))
	}

	if len(e.Details) > 0 {
		b.WriteString(fmt.Sprintf(", details=%v", e.Details))
	}

	return b.String()
}
