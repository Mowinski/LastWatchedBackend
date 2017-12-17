package utils

import "testing"

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
