package render

import (
	"html/template"
	"net/http"
)

type HTML struct {
	Name   string
	Data   any
	Temp   *template.Template
	IsTemp bool
}

func (H *HTML) Render(writer http.ResponseWriter, status int) error {
	H.WriteContentType(writer)
	writer.WriteHeader(status)
	if H.IsTemp {
		return H.Temp.ExecuteTemplate(writer, H.Name, H.Data)
	}
	_, err := writer.Write([]byte(H.Data.(string)))
	return err
}

func (H *HTML) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
}

type HTMLRender struct {
	Template *template.Template
	Name     string
	Data     any
}
