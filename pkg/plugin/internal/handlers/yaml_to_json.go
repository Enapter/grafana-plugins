package handlers

import (
	"encoding/json"
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

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
