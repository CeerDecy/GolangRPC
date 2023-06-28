package crpcLogger

import (
	"encoding/json"
	"time"
)

const jsonFormatterTag = "JsonFormatter"

type JsonFormatter struct {
}

func (j *JsonFormatter) Format(param *LoggerFormatterParam) string {
	jsonMap := make(map[string]any)
	jsonMap["log_time"] = time.Now().Format("2006/01/02 - 15:04:05")
	jsonMap["msg"] = param.Msg
	jsonMap["TAG"] = param.Tag
	if param.Fields != nil {
		jsonMap["data"] = param.Fields
	}
	bytes, err := json.Marshal(jsonMap)
	if err != nil {
		logger := TextLogger()
		logger.Error(jsonFormatterTag, err)
		return ""
	}
	return string(bytes)
}
