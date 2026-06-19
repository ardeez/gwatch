package logger

import (
	"fmt"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

func (l Level) stringRepresentation() (string, string) {
	switch l {
	case DEBUG:
		return "DEBUG", colorCyan
	case INFO:
		return "INFO ", colorGreen
	case WARN:
		return "WARN ", colorYellow
	case ERROR:
		return "ERROR", colorRed
	default:
		return "LOG  ", colorReset
	}
}

func logFormat(level Level, format string, v ...interface{}) {
	levelStr, color := level.stringRepresentation()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, v...)
	fmt.Printf("%s[%s]%s %s[%s]%s %s\n",
		colorBlue, timestamp, colorReset,
		color, levelStr, colorReset,
		msg,
	)
}

func Debug(format string, v ...interface{}) { logFormat(DEBUG, format, v...) }
func Info(format string, v ...interface{})  { logFormat(INFO, format, v...) }
func Warn(format string, v ...interface{})  { logFormat(WARN, format, v...) }
func Error(format string, v ...interface{}) { logFormat(ERROR, format, v...) }
