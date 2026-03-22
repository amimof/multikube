// Package logger provides interfaces and implementations for working with logs
package logger

import (
	"log"
	"runtime"
	"strings"
)

type Logger interface {
	Debug(string, ...any)
	Info(string, ...any)
	Warn(string, ...any)
	Error(string, ...any)
}

type NilLogger struct{}

func (c NilLogger) Debug(msg string, fields ...any) {}
func (c NilLogger) Info(msg string, fields ...any)  {}
func (c NilLogger) Warn(msg string, fields ...any)  {}
func (c NilLogger) Error(msg string, fields ...any) {}

type ConsoleLogger struct{}

func (c ConsoleLogger) Debug(msg string, fields ...any) {
	c.log("DEBUG", msg, fields...)
}

func (c ConsoleLogger) Info(msg string, fields ...any) {
	c.log("INFO", msg, fields...)
}

func (c ConsoleLogger) Warn(msg string, fields ...any) {
	c.log("WARN", msg, fields...)
}

func (c ConsoleLogger) Error(msg string, fields ...any) {
	c.log("ERROR", msg, fields...)
}

func (c *ConsoleLogger) log(level, msg string, fields ...any) {
	file, line, funcName := getCallerInfo(2)
	log.Printf("[%s] %s:%d %s() - "+msg, append([]any{level, file, line, funcName}, fields...)...)
}

// getCallerInfo gets the file, line, and function name of the caller
func getCallerInfo(skip int) (string, int, string) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown_file", 0, "unknown_func"
	}

	// Get function name
	funcName := runtime.FuncForPC(pc).Name()
	funcName = trimFunctionName(funcName)

	// Trim the file path to only the base name
	fileParts := strings.Split(file, "/")
	file = fileParts[len(fileParts)-1]

	return file, line, funcName
}

func trimFunctionName(funcName string) string {
	funcParts := strings.Split(funcName, "/")
	return funcParts[len(funcParts)-1]
}

var _ Logger = &DevNullLogger{}

type DevNullLogger struct{}

// Debug implements Logger.
func (d *DevNullLogger) Debug(string, ...any) {
}

// Error implements Logger.
func (d *DevNullLogger) Error(string, ...any) {
}

// Info implements Logger.
func (d *DevNullLogger) Info(string, ...any) {
}

// Warn implements Logger.
func (d *DevNullLogger) Warn(string, ...any) {
}
