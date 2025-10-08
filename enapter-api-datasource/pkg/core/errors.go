package core

import (
	"errors"
	"fmt"
	"strings"
)

var ErrTimeseriesEmpty = errors.New("timeseries empty")

type EnapterAPIError struct {
	Code    string
	Message string
	Details map[string]any
}

func (e EnapterAPIError) Error() string {
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
