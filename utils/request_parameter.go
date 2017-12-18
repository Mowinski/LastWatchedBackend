package utils

import (
	"encoding/json"
	"io"
	"strconv"
)

// GetIntOrDefault return value as string or if value is empty or not string return defaultValue
func GetIntOrDefault(value string, defaultValue int) int {
	if len(value) == 0 {
		return defaultValue
	}
	ret, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return ret
}

// GetJSONParameters ...
func GetJSONParameters(body io.ReadCloser, out interface{}) error {
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&out)
	defer body.Close()
	return err
}
