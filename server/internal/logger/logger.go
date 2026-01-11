package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[90m"
)

var fileWriter io.Writer

func SetFile(w io.Writer) {
	fileWriter = w
}

func log(color, prefix, msg string) {
	ts := time.Now().Format("15:04:05")
	colored := fmt.Sprintf("%s%s%s %s%s%s %s", Gray, ts, Reset, color, prefix, Reset, msg)
	plain := fmt.Sprintf("%s %s %s", ts, prefix, msg)

	fmt.Println(colored)
	if fileWriter != nil {
		fmt.Fprintln(fileWriter, plain)
	}
}

func Info(format string, args ...interface{}) {
	log(Cyan, "INFO", fmt.Sprintf(format, args...))
}

func Success(format string, args ...interface{}) {
	log(Green, "OK", fmt.Sprintf(format, args...))
}

func Warn(format string, args ...interface{}) {
	log(Yellow, "WARN", fmt.Sprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	log(Red, "ERROR", fmt.Sprintf(format, args...))
}

func Plugin(pluginID, msg string) {
	ts := time.Now().Format("15:04:05")
	colored := fmt.Sprintf("%s%s%s %s[%s]%s %s", Gray, ts, Reset, Magenta, pluginID, Reset, msg)
	plain := fmt.Sprintf("%s [%s] %s", ts, pluginID, msg)

	fmt.Println(colored)
	if fileWriter != nil {
		fmt.Fprintln(fileWriter, plain)
	}
}

func Fatal(format string, args ...interface{}) {
	log(Red, "FATAL", fmt.Sprintf(format, args...))
	os.Exit(1)
}

type StdLogger struct{}

func (l *StdLogger) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if strings.Contains(msg, "[plugin:") {
		start := strings.Index(msg, "[plugin:") + 8
		end := strings.Index(msg[start:], "]")
		if end > 0 {
			pluginID := msg[start : start+end]
			rest := strings.TrimSpace(msg[start+end+1:])
			Plugin(pluginID, rest)
			return len(p), nil
		}
	}
	Info("%s", msg)
	return len(p), nil
}

func NewStdLogger() *StdLogger {
	return &StdLogger{}
}
