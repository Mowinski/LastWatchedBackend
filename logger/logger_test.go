package logger

import (
	"os"
	"testing"
)

func TestLoggerCanSet(t *testing.T) {
	const TestLogFile = "test_file.log"
	const TestLogFileNotExist = "/file/not/exists"

	if Logger != nil {
		t.Error("Logger is setting without initialization")
	}
	err := SetLogger(TestLogFile)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if Logger == nil {
		t.Error("Logger is nil, expected initialized object")
	}

	SetLogger(TestLogFileNotExist)
	if Logger != nil {
		t.Error("Logger is initialized but error occure, expected nil")
	}

	os.Remove(TestLogFile)
}
