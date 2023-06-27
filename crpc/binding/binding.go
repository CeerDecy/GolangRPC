package binding

import "net/http"

type Binding interface {
	Name() string
	Bind(*http.Request, any) error
}

var JSON Binding = &jsonBinding{}
var XML Binding = &xmlBinding{}
