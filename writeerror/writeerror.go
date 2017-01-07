// Package writeerror is used to configure how errors are
// marshalled and logged in calls to httpapi.WriteError.
//
// This has been put in a separate package to reduce the surface area
// of the httpapi package API. This package is called when setting
// up the Web API server middleware, where the httpapi package is referenced
// in HTTP handlers.
package writeerror

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// Content contains the information sent back to the HTTP client
// in an error response.
type Content struct {
	Message string // Message sent to client, which may be different to err.Error().
	Status  int    // HTTP status
	Code    string // Optional Error code
	Trace   string // Optional unique ID for cross reference with tracing/logging
	Err     error  // Only sent to trusted clients
}

// Config contains configuration in the form of callback functions that are
// called during calls to httpapi.WriteError.
type Config struct {
	// GetTrace specifies an optional callback function that returns an identifier
	// for correlating the error response with a trace or a set of log entries. The
	// default implementation returns an empty string.
	GetTrace func(*http.Request) string

	// IsTrusted specifies an optional callback function
	// that is called to determine if the client is trusted
	// to receive error messages that include implementation
	// details. By default a client is trusted if it is invoked
	// from the local host without a reverse proxy.
	IsTrusted func(*http.Request) bool

	// MarshalContentCallback specifies an optional callback function
	// that is called to marshal error details into JSON. If not specified
	// an error is marshalled into the following JSON:
	//  {
	//      "error": {
	//          "message": "message text",
	//          "status": 400,
	//          "code": "XXX999",
	//          "trace": "a8845f4dc3792a63",
	//          "detail": "detailed information for trusted clients"
	//      }
	//  }
	// In the example above, the "code", "trace" and "detail" keys are optional.
	MarshalContent func(*Content) []byte

	// ErrorWrittenCallback specifies an optional callback function that is called whenever
	// an error has been written to the client. This can be used to log all error
	// messages sent to the client. The default implementation logs to the standard
	// logger.
	ErrorWritten func(*http.Request, *Content)
}

// Default contains the default configuration callbacks.
var Default Config

func init() {
	Default.GetTrace = defaultGetTrace
	Default.IsTrusted = defaultIsTrusted
	Default.MarshalContent = defaultMarshalContent
	Default.ErrorWritten = defaultErrorWritten
}

type contextKey int

// Keys for storing values in the context.
const (
	errorCallbackKey contextKey = 0
)

// newRequest associates the error callbacks with the current request, returning
// a request with a new context.
//
// This function is intended to be called from HTTP request middleware.
func (c Config) newRequest(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), errorCallbackKey, c)
	return r.WithContext(ctx)
}

// ConfigFromRequest extracts the config from the HTTP request. If Middleware
// was used to insert a config then that config will be returned. Otherwise the
// default configuration is used.
//
// The Config returned by this function will always have non-nil values for all
// callbacks, pointing to the default implementation if not specified otherwise.
func ConfigFromRequest(r *http.Request) Config {
	config, _ := r.Context().Value(errorCallbackKey).(Config)
	if config.GetTrace == nil {
		config.GetTrace = Default.GetTrace
	}
	if config.IsTrusted == nil {
		config.IsTrusted = Default.IsTrusted
	}
	if config.MarshalContent == nil {
		config.MarshalContent = Default.MarshalContent
	}
	if config.ErrorWritten == nil {
		config.ErrorWritten = Default.ErrorWritten
	}
	return config
}

// Middleware returns middleware that associates the Callback
// with the HTTP request. Use this in the middleware stack to customise how
// errors are marshalled and reported.
func Middleware(c Config) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = c.newRequest(r)
			h.ServeHTTP(w, r)
		})
	}
}

func defaultGetTrace(r *http.Request) string {
	return ""
}

func defaultIsTrusted(r *http.Request) bool {
	// TODO(jpj): check for localhost request
	return false
}

func defaultMarshalContent(content *Content) []byte {
	var payload struct {
		Error struct {
			Message string `json:"message"`
			Status  int    `json:"status"`
			Code    string `json:"code,omitempty"`
			Trace   string `json:"trace,omitempty"`
			Detail  string `json:"detail,omitempty"`
		} `json:"error"`
	}
	payload.Error.Message = content.Message
	payload.Error.Status = content.Status
	payload.Error.Code = content.Code
	payload.Error.Trace = content.Trace
	if content.Err != nil {
		payload.Error.Detail = content.Err.Error()
	}

	// format errors nicely to make diagnostics easier when using curl
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(payload) // should not fail

	return buf.Bytes()
}

func defaultErrorWritten(r *http.Request, content *Content) {

}
