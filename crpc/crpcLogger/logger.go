package crpcLogger

import (
	"fmt"
	"io"
	"log"
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

// WriterInFile 设置将文件写入对应文件中
func (logger *Logger) WriterInFile(filepath string) {
	logger.AddWriter(FileWriter(filepath))
}

// AddWriter 添加Writer
func (logger *Logger) AddWriter(writer io.Writer) {
	logger.Outs = append(logger.Outs, writer)
}

type LoggerFormatter interface {
	Format(param *LoggerFormatterParam) string
}

type LoggerFormatterParam struct {
	Level   LogLevel
	IsColor bool
	Tag     string
	Msg     any
	Fields  map[string]any
}

//func Default() *Logger {
//	logger := NewLogger()
//	logger.Level = LevelDebug
//	logger.Outs = append(logger.Outs, os.Stdout)
//	logger.Formatter = &TextFormatter{}
//	return logger
//}

func TextLogger() *Logger {
	logger := NewLogger()
	logger.Level = LevelDebug
	logger.Outs = append(logger.Outs, os.Stdout)
	logger.Formatter = &TextFormatter{}
	return logger
}

func JsonLogger() *Logger {
	logger := NewLogger()
	logger.Level = LevelDebug
	logger.Outs = append(logger.Outs, os.Stdout)
	logger.Formatter = &JsonFormatter{}
	return logger
}

func NewLogger() *Logger {
	return &Logger{}
}

func (logger *Logger) InfoFields(tag string, msg any, fields map[string]any) {
	logger.Print(LevelInfo, tag, msg, fields)
}

func (logger *Logger) Info(tag string, msg any) {
	logger.Print(LevelInfo, tag, msg, nil)
}

func (logger *Logger) DebugFields(tag string, msg any, fields map[string]any) {
	logger.Print(LevelDebug, tag, msg, fields)
}

func (logger *Logger) Debug(tag string, msg any) {
	logger.Print(LevelDebug, tag, msg, nil)
}

func (logger *Logger) ErrorFields(tag string, msg any, fields map[string]any) {
	logger.Print(LevelError, tag, msg, fields)
}

func (logger *Logger) Error(tag string, msg any) {
	logger.Print(LevelError, tag, msg, nil)
}

func (logger *Logger) Print(level LogLevel, tag string, msg any, fields map[string]any) {
	if logger.Level > level {
		// 当前级别大于输入级别，则不打印
		return
	}
	param := &LoggerFormatterParam{
		Level:  level,
		Tag:    tag,
		Msg:    msg,
		Fields: fields,
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

func FileWriter(name string) io.Writer {
	writer, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Panicln(err)
	}
	return writer
}
