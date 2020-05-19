package log

import "fmt"

var defaultLogger = &Logger{}

type Logger struct {
}

func (l *Logger) print(level string, format string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Println(fmt.Sprintf(level+format, args...))
		return
	}
	fmt.Println(level + format)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.print("[INFO] ", format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.print("[Warn] ", format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.print("[EROR] ", format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}
