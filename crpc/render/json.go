package render

import (
	"encoding/json"
	"net/http"
)

type Json struct {
	Data any
}

func (j *Json) Render(writer http.ResponseWriter) error {
	j.WriteContentType(writer)
	data, err := json.Marshal(j.Data)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

func (j *Json) WriteContentType(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/json")
}
