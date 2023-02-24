package telemetryapi

import (
	"fmt"
	"io"
)

func dumpBody(body io.Reader) (string, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("read: %w", err)
	}

	if len(data) == 0 {
		return "", errEmptyData
	}

	const maxSize = 200

	if len(data) < maxSize {
		return string(data), nil
	}

	dump := fmt.Sprintf("%s[...] (full len = %d)", data[:maxSize], len(data))

	return dump, nil
}
