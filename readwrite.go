package httpapi

import (
	"fmt"
	"net/http"

	"github.com/jjeffery/errkind"
	"github.com/jjeffery/errors"
	"github.com/jjeffery/httpapi/writeerror"
)

// ReadRequest reads the request body as JSON, and unmarshals it
// into the structure pointed to by body.
//
// Although not specified in the HTTP spec, if the request contains a
// header "Content-Encoding: gzip", then the request body will be decompressed.
// This is convenient for HTTP clients that PUT or POST large JSON content.
func ReadRequest(r *http.Request, body interface{}) error {
	var data rawData
	if err := data.ReadRequest(r); err != nil {
		return err
	}
	if err := data.UnmarshalTo(body); err != nil {
		return err
	}
	return nil
}

// WriteResponse sends the response as JSON to the HTTP client. The
// response is compressed if the HTTP client is able to accept compressed
// responses.
func WriteResponse(w http.ResponseWriter, r *http.Request, body interface{}) {
	// Special case if the body is an error.
	if err, ok := body.(error); ok {
		WriteError(w, r, err)
		return
	}

	var data rawData

	if err := data.MarshalFrom(body); err != nil {
		WriteError(w, r, err)
		return
	}

	if err := data.CompressResponse(r); err != nil {
		WriteError(w, r, err)
		return
	}

	// TODO(jpj): log this if  logging/tracing becomes available
	_ = data.WriteResponse(w)
}

// WriteError writes an error message as a JSON object.
//
// The HTTP status code is retrieved from the error using
// the errkind package. If no status is associated with the
// error then a 500 status is returned.
//
// Care is taken to ensure that no implementation details are
// leaked to the client. If the error implements the `publicer` interface
// (as defined in the errkind package), then the error's message and
// HTTP status are considered suitable for returning to the client.
// Otherwise a more general error message is returned.
//
// If the client is considered a "trusted" client, then full details
// of the error are returned in an extra details key that is not present
// for untrusted clients. By default a client is trusted if the request
// originated on the local host (but does not include requests routed through
// a local reverse proxy).
//
// The writeerror subdirectory package provides configuration on how errors are marshalled
// to the client, and how details of the errors are logged and/or traced. The
// defaults are sensible, so this function can be used with no configuration.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		err = errkind.Public("no information available", http.StatusInternalServerError)
	}
	config := writeerror.ConfigFromRequest(r)

	// build the content to send to the client
	var content writeerror.Content
	{
		cause := errors.Cause(err)

		// use the status code if it is public
		if _, ok := cause.(interface{ PublicStatusCode() }); ok {
			content.StatusCode = errkind.StatusCode(cause)
		}
		if content.StatusCode < 400 || content.StatusCode > 599 {
			content.StatusCode = http.StatusInternalServerError
		}

		// use the message if it is public, otherwise use the
		// message for the status code
		if _, ok := cause.(interface{ PublicMessage() }); ok {
			// The errkind package has errors that have a Message() method
			// that returns the message without the code. Useful here because
			// the code is kept in a separate field in the returned error.
			// TODO(jpj): this seems a little overcomplicated.
			if messager, ok := cause.(interface{ Message() string }); ok {
				content.Message = messager.Message()
			} else {
				content.Message = cause.Error()
			}
		}
		if content.Message == "" {
			content.Message = http.StatusText(content.StatusCode)
		}

		if _, ok := cause.(interface{ PublicCode() }); ok {
			content.Code = errkind.Code(cause)
		}

		content.Trace = config.GetTrace(r)

		if config.IsTrusted(r) {
			// only include the error in the content for trusted clients
			content.Err = err
		}
	}

	// build the content bytes to write to the client
	data := config.MarshalContent(&content)

	// write the response to the client
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(content.StatusCode)
	w.Write(data)

	// Populate the Err property if it has not been populated earlier
	// so that it can be included in log messages or other diagnostics.
	content.Err = err

	// call errorWritten for logging/tracing/diagnostics
	config.ErrorWritten(r, &content)
}
