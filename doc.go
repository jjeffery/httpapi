// Copyright (c) 2016 John Jeffery. All rights reserved.

// Package httpapi provides assistance for servers that implement a JSON Web API.
// It has no pretensions to be a framework: it is intended to work with the standard
// library HTTP handlers. The goal is to provide simple primitives for reading
// input from HTTP requests and writing output to the HTTP response writer.
//
// The package supports compressing responses if the client can support it. It also
// provides non-standard support for clients compressing the body of requests. See
// the ReadRequest function for more details.
//
// The WriteError function provides a simple, consistent way to send error messages
// to HTTP clients. It has some sensible defaults, but these may not suit. The behaviour
// of the WriteError function can be customized: see the writeerror subdirectory package.
package httpapi
