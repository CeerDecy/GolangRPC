package render

import (
	"fmt"
	"net/http"
)

type String struct {
	Format string
	Data   []any
}

func (str *String) Render(w http.ResponseWriter) error {
	str.WriteContentType(w)
	if len(str.Data) > 0 {
		_, err := fmt.Fprintf(w, str.Format, str.Data...)
		return err
	}
	_, err := w.Write([]byte(str.Format))
	return err
}
func (str *String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/plain; charset=utf-8")
}
