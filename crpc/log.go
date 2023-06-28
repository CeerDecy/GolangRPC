package crpc

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

type LoggingConfig struct {
	Formatter LoggerFormatter
	out       io.Writer
}

var DefaultWriter io.Writer = os.Stdout

type LoggerFormatter = func(params *LogFormatterParams) string

var defaultFormatter = func(params *LogFormatterParams) string {
	var codeColor = params.StateCodeColor()
	var resetColor = params.ResetColor()
	var methodColor = params.MethodColor()
	if params.Latency > time.Minute {
		params.Latency = params.Latency.Truncate(time.Second)
	}
	if params.IsDisplayColor {
		return fmt.Sprintf("[crpc] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v",
			params.TimeStamp.Format("2006/01/02 - 15:04:05"),
			codeColor, params.Code, resetColor,
			params.Latency, params.ClientIP,
			methodColor, params.Method, resetColor,
			params.Path,
		)
	} else {
		return fmt.Sprintf("[crpc] %v | %3d | %13v | %15s | %-7s %#v",
			params.TimeStamp.Format("2006/01/02 - 15:04:05"),
			params.Code,
			params.Latency,
			params.ClientIP,
			params.Method,
			params.Path,
		)
	}

}

type LogFormatterParams struct {
	Request        *http.Request
	TimeStamp      time.Time
	Code           int
	Latency        time.Duration
	ClientIP       net.IP
	Method         string
	Path           string
	IsDisplayColor bool
}

func (p *LogFormatterParams) StateCodeColor() string {
	code := p.Code
	switch code {
	case http.StatusOK:
		return green
	case http.StatusNotFound:
		return yellow
	default:
		return red
	}
}

func (p *LogFormatterParams) ResetColor() string {
	return reset
}

func (p *LogFormatterParams) MethodColor() string {
	switch p.Method {
	case http.MethodGet:
		return green
	case http.MethodPost:
		return yellow
	case http.MethodPut:
		return blue
	case http.MethodDelete:
		return red
	default:
		return white
	}
}

// LoggingWithConfig 打印日志
func LoggingWithConfig(conf LoggingConfig, next HandleFunc) HandleFunc {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultFormatter
	}
	out := conf.out
	if out == nil {
		out = DefaultWriter
	}
	return func(ctx *Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		raw := ctx.Request.URL.RawQuery
		next(ctx)
		end := time.Now()
		latency := end.Sub(start)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(ctx.Request.RemoteAddr))
		clientIP := net.ParseIP(ip)
		method := ctx.Request.Method

		if raw != "" {
			path = path + "?" + raw
		}
		params := &LogFormatterParams{
			ctx.Request,
			end,
			ctx.code,
			latency,
			clientIP,
			method,
			path,
			true,
		}
		_, _ = fmt.Fprintln(out, formatter(params))
	}
}

func Logging(next HandleFunc) HandleFunc {
	return LoggingWithConfig(LoggingConfig{}, next)
}
