package binding

import (
	"encoding/xml"
	"errors"
	"net/http"
)

type xmlBinding struct {
}

func (x xmlBinding) Name() string {
	return "xml"
}

func (x xmlBinding) Bind(request *http.Request, model any) error {
	if request.Body == nil {
		return errors.New("request Body is nil")
	}
	decoder := xml.NewDecoder(request.Body)
	err := decoder.Decode(model)
	if err != nil {
		return err
	}
	return validate(model)
}
