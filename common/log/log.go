package log

import (
	"fmt"
	"log"
	"time"
)

func Info(format string, args ...interface{}) {
	printf("INFO", format, args...)
}

func Warn(format string, args ...interface{}) {
	printf("WARN", format, args...)
}

func Error(format string, args ...interface{}) {
	printf("EROR", format, args...)
}

func printf(level, format string, args ...interface{}) {
	t := time.Now().Format("2006-01-02 15:04:05")
	log.Printf(fmt.Sprintf("[%s] %s %s", level, t, format), args...)
}
