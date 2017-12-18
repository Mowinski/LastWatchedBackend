package utils

import (
	"testing"
)

type JSONMock struct {
	Name string
}
type mockedBody string

func (m mockedBody) Read(p []byte) (n int, err error) {
	copy(p, m)
	return len(m), nil
}
func (m mockedBody) Close() error {
	return nil
}
func TestGetIntOrDefault(t *testing.T) {
	var value int

	value = GetIntOrDefault("15", 0)
	if value != 15 {
		t.Errorf("GetIntOrDefault return %d, expected 15, correct value", value)
	}

	value = GetIntOrDefault("NotInt", 15)
	if value != 15 {
		t.Errorf("GetIntOrDefault return %d, expected 15, incorrect value", value)
	}

	value = GetIntOrDefault("", 15)
	if value != 15 {
		t.Errorf("GetIntOrDefault return %d, expected 15, empty value", value)
	}
}

func TestGetJSONParameters(t *testing.T) {
	var body mockedBody = "{ \"name\": \"test\" }"
	var out JSONMock
	GetJSONParameters(body, &out)

	if out.Name != "test" {
		t.Errorf("Wrong return value, expected 'test', got %s", out.Name)
	}
}
