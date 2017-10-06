package httpapi

import "net/http"

// HandlerFunc is similar to http.HandlerFunc, but it returns an error.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// ServeHTTP implements the http.Handler interface.
func (fn HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		WriteError(w, r, err)
	}
}
