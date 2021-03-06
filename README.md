# httpapi [![GoDoc](https://godoc.org/github.com/jjeffery/httpapi?status.svg)](https://godoc.org/github.com/jjeffery/httpapi) [![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/jjeffery/httpapi/master/LICENSE.md) [![Build Status](https://travis-ci.org/jjeffery/httpapi.svg?branch=master)](https://travis-ci.org/jjeffery/httpapi) [![Coverage Status](https://coveralls.io/repos/github/jjeffery/httpapi/badge.svg?branch=master)](https://coveralls.io/github/jjeffery/httpapi?branch=master) [![GoReportCard](https://goreportcard.com/badge/github.com/jjeffery/httpapi)](https://goreportcard.com/report/github.com/jjeffery/httpapi)

Package `httpapi` provides support for implementing HTTP servers that expose a JSON API.

Example of a handler that extracts the input from the body of the HTTP request.
```go
func postHandler(w http.ResponseWriter, r *http.Request) {
    // unmarshal input from request
    var input DoSomethingInput
    if err := httpapi.ReadRequest(w, r, &input); err != nil {
        httpapi.WriteError(w, r, err)
        return
    }

    output, err := doSomethingWith(r.Context(), &input)
    if err != nil {
        httpapi.WriteError(w, r, err)
        return
    }

    httpapi.WriteResponse(w, r, output)
}
```

Example of a handler that extracts the input from the query strings of the HTTP request.
```go
func getHandler(w http.ResponseWriter, r *http.Request) {
    query := httpapi.Query(r)
    input := DoAnotherThingInput {
        Since: query.GetTime("since"),
        Limit: query.GetInt("limit"),
        Offset: query.GetInt("offset"),
    }

    // validate once after all query string parameters have been read
    if err := query.Err(); err != nil {
        httpapi.WriteResponse(w, r, err)
        return
    }

    output, err := doAnotherThingWith(r.Context(), &input)
    if err != nil {
        httpapi.WriteError(w, r, err)
        return
    }

    httpapi.WriteResponse(w, r, output)
}
```

[Read the package documentation for more information](https://godoc.org/github.com/jjeffery/httpapi).
