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

func (H *HTML) Render(w http.ResponseWriter) error {
	H.WriteContentType(w)
	if H.IsTemp {
		return H.Temp.ExecuteTemplate(w, H.Name, H.Data)
	}
	_, err := w.Write([]byte(H.Data.(string)))
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
