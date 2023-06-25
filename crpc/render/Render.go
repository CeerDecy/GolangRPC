package render

import "net/http"

type Render interface {
	Render(w http.ResponseWriter) error
	WriteContentType(w http.ResponseWriter)
}

func writeContentType(writer http.ResponseWriter, value string) {
	writer.Header().Set("Content-Type", value)
}
