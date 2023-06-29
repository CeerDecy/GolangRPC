package render

import (
	"fmt"
	"net/http"
)

type String struct {
	Format string
	Data   []any
}

func (str *String) Render(writer http.ResponseWriter, status int) error {
	str.WriteContentType(writer)
	writer.WriteHeader(status)
	if len(str.Data) > 0 {
		_, err := fmt.Fprintf(writer, str.Format, str.Data...)
		return err
	}
	_, err := writer.Write([]byte(str.Format))
	return err
}
func (str *String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/plain; charset=utf-8")
}
