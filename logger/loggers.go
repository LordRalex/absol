package logger

import (
	"io"
	"log"
	"os"
)

var errorLogger *log.Logger
var outLogger *log.Logger
var debugLogger *log.Logger
var logFile *os.File

func init() {
	var err error
	logFile, err = os.OpenFile("output.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Printf("Error loading log file: %s", err.Error())
	}

	var output io.Writer
	var errorOut io.Writer
	var debugOut io.Writer
	if logFile != nil {
		output = io.MultiWriter(os.Stdout, logFile)
		errorOut = io.MultiWriter(os.Stderr, logFile)
		debugOut = io.MultiWriter(os.Stdout, logFile)
	} else {
		output = os.Stdout
		errorOut = os.Stderr
		debugOut = os.Stdout
	}

	errorLogger = log.New(errorOut, "[ERROR] ", log.Flags())
	outLogger = log.New(output, "[INFO] ", log.Flags())
	debugLogger = log.New(debugOut, "[DEBUG] ", log.Flags())
}

func Close() error {
	return logFile.Close()
}

func Out() *log.Logger {
	return outLogger
}

func Err() *log.Logger {
	return errorLogger
}

func Debug() *log.Logger {
	return debugLogger
}
