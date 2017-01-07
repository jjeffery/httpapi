package httpapi_test

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jjeffery/httpapi"
)

func Example() {
	r := mux.NewRouter()
	r.Path("/api/something").Methods("POST").HandlerFunc(postHandler)
	r.Path("/api/something").Methods("GET").HandlerFunc(getHandler)
	http.ListenAndServe(":8080", r)
}

// postHandler handles POST requests
func postHandler(w http.ResponseWriter, r *http.Request) {
	// unmarshal input from request
	var input PostSomethingInput
	if err := httpapi.ReadRequest(r, &input); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}

	output, err := postSomething(r.Context(), &input)
	if err != nil {
		httpapi.WriteError(w, r, err)
		return
	}

	httpapi.WriteResponse(w, r, output)
}

// getHandler extracts the input from the query strings of the HTTP request.
func getHandler(w http.ResponseWriter, r *http.Request) {
	query := httpapi.Query(r)
	input := GetSomethingInput{
		Search: query.GetString("q"),
		Since:  query.GetTime("since"),
		Limit:  query.GetInt("limit"),
		Offset: query.GetInt("offset"),
	}

	// validate once after all query string parameters have been read
	if err := query.Err(); err != nil {
		httpapi.WriteResponse(w, r, err)
		return
	}

	output, err := getSomething(r.Context(), &input)
	if err != nil {
		httpapi.WriteError(w, r, err)
		return
	}

	httpapi.WriteResponse(w, r, output)
}
