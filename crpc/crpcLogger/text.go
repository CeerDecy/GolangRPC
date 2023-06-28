package crpcLogger

import (
	"fmt"
	"time"
)

type TextFormatter struct {
}

// Level 获取等级名称
func (t *TextFormatter) Level(level LogLevel) string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	default:
		return ""
	}
}

// LevelColor 获取对应等级颜色
func (t *TextFormatter) LevelColor(level LogLevel) string {
	switch level {
	case LevelDebug:
		return yellow
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return white
	}
}

// Format 文本日志格式化
func (t *TextFormatter) Format(param *LoggerFormatterParam) string {
	now := time.Now()
	if param.IsColor {
		return fmt.Sprintf("[crpc] %v | %s%s%s | [ %s ]:%s",
			now.Format("2006/01/02 - 15:04:05"),
			t.LevelColor(param.Level), t.Level(param.Level), reset,
			param.Tag, param.Msg,
		)
	} else {
		return fmt.Sprintf("[crpc] %v | %s | [ %s ]:%s",
			now.Format("2006/01/02 - 15:04:05"),
			t.Level(param.Level),
			param.Tag, param.Msg,
		)
	}
}
