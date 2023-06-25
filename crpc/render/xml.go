package render

import (
	"encoding/xml"
	"net/http"
)

type XML struct {
	Data any
}

func (X *XML) Render(w http.ResponseWriter) error {
	X.WriteContentType(w)
	return xml.NewEncoder(w).Encode(X.Data)

}

func (X *XML) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/xml")
}
