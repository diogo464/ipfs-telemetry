package telemetry

import (
	"encoding/json"
)

func JsonStreamDecoder(encoded []byte) (string, error) {
	var obj map[string]interface{}
	err := json.Unmarshal(encoded, &obj)
	if err != nil {
		return "", err
	}
	ndjson, err := json.Marshal(obj)
	return string(ndjson), nil
}

func JsonPrettyStreamDecoder(encoded []byte) (string, error) {
	var obj map[string]interface{}
	err := json.Unmarshal(encoded, &obj)
	if err != nil {
		return "", err
	}
	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(pretty), nil
}
