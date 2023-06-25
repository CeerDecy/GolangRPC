package render

import (
	"errors"
	"fmt"
	"net/http"
)

type Redirect struct {
	Code     int
	Req      *http.Request
	Location string
}

func (r *Redirect) Render(w http.ResponseWriter) error {
	if r.Code < http.StatusMultipleChoices || r.Code > http.StatusPermanentRedirect && r.Code != http.StatusCreated {
		return errors.New(fmt.Sprintf("Cannot redirect with code %d", r.Code))
	}
	http.Redirect(w, r.Req, r.Location, r.Code)
	return nil
}

func (r *Redirect) WriteContentType(w http.ResponseWriter) {
}
