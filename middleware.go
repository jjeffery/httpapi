package httpapi

import "net/http"

// Middleware is a function that filters a request coming into
// the application and responses going back to the client. Middleware
// is similar in concept to Rack middleware.
type Middleware func(http.Handler) http.Handler

// A Stack is a stack of middleware functions that are common to one or more
// HTTP handlers. A middleware function is any function that accepts a Handler as a
// parameter and returns a Handler.
type Stack struct {
	middleware Middleware
	previous   *Stack
}

// Use creates a Stack of middleware functions.
func Use(f ...Middleware) *Stack {
	var stack *Stack

	for _, m := range f {
		if m != nil {
			stack = &Stack{
				middleware: m,
				previous:   stack,
			}
		}
	}

	return stack
}

// Use creates a new stack by appending the middleware functions to
// the existing stack.
func (s *Stack) Use(f ...Middleware) *Stack {
	stack := s

	for _, m := range f {
		if m != nil {
			stack = &Stack{
				middleware: m,
				previous:   stack,
			}
		}
	}

	return stack
}

// Handler creates a http.Handler from a stack of middleware
// functions and a httpctx.Handler.
func (s *Stack) Handler(h http.Handler) http.Handler {
	for stack := s; stack != nil; stack = stack.previous {
		if stack.middleware != nil {
			h = stack.middleware(h)
		}
	}

	return h
}

// HandlerFunc returns a http.Handler (compatible with the standard library http package), which
// calls the middleware handlers in the stack s, followed by  the handler function f.
func (s *Stack) HandlerFunc(f func(http.ResponseWriter, *http.Request)) http.Handler {
	if s == nil {
		return http.HandlerFunc(f)
	}
	return s.Handler(http.HandlerFunc(f))
}
