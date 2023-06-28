package crpcLogger

import (
	"fmt"
	"io"
	"os"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelError
)

type Logger struct {
	Level     LogLevel
	Outs      []io.Writer
	Formatter LoggerFormatter
}

type LoggerFormatter interface {
	Format(param *LoggerFormatterParam) string
}

type LoggerFormatterParam struct {
	Level   LogLevel
	IsColor bool
	Tag     string
	Msg     any
}

func Default() *Logger {
	logger := NewLogger()
	logger.Level = LevelDebug
	logger.Outs = append(logger.Outs, os.Stdout)
	logger.Formatter = &TextFormatter{}
	return logger
}

func NewLogger() *Logger {
	return &Logger{}
}

func (logger *Logger) Info(tag string, msg any) {
	logger.Print(LevelInfo, tag, msg)
}
func (logger *Logger) Debug(tag string, msg any) {
	logger.Print(LevelDebug, tag, msg)

}
func (logger *Logger) Error(tag string, msg any) {
	logger.Print(LevelError, tag, msg)

}

func (logger *Logger) Print(level LogLevel, tag string, msg any) {
	if logger.Level > level {
		// 当前级别大于输入级别，则不打印
		return
	}
	param := &LoggerFormatterParam{
		Level: level,
		Tag:   tag,
		Msg:   msg,
	}
	param.Level = level
	commonStr := logger.Formatter.Format(param)
	for _, out := range logger.Outs {
		if out == os.Stdout {
			param.IsColor = true
			colorStr := logger.Formatter.Format(param)
			_, _ = fmt.Fprintln(out, colorStr)
		} else {
			_, _ = fmt.Fprintln(out, commonStr)
		}
	}
}
