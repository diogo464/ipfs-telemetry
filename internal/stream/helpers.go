package stream

import (
	"encoding/json"
)

func JsonObjDecoder(encoded []byte) (map[string]interface{}, error) {
	var obj map[string]interface{}
	err := json.Unmarshal(encoded, &obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func JsonDecoder(encoded []byte) (string, error) {
	var obj map[string]interface{}
	err := json.Unmarshal(encoded, &obj)
	if err != nil {
		return "", err
	}
	ndjson, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(ndjson), nil
}

func JsonPrettyDecoder(encoded []byte) (string, error) {
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
