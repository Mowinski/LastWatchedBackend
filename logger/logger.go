package logger

import (
	"io"
	"log"
	"os"
)

// Logger is a logger object which print log into file and stdout
var Logger *log.Logger

// SetLogger create new logger object and associate it with global Logger variable
func SetLogger(logFileName string) (err error) {
	outputFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		Logger = nil
		return err
	}
	Logger = log.New(outputFile, "", log.Lshortfile)
	Logger.SetOutput(io.MultiWriter(os.Stdout, outputFile))
	return nil
}
