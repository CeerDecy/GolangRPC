package render

import (
	"encoding/xml"
	"net/http"
)

type XML struct {
	Data any
}

func (X *XML) Render(writer http.ResponseWriter, status int) error {
	X.WriteContentType(writer)
	writer.WriteHeader(status)
	return xml.NewEncoder(writer).Encode(X.Data)
}

func (X *XML) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/xml")
}
