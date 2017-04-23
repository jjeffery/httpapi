// Copyright 2016 John Jeffery <john@jeffery.id.au>. All rights reserved.

package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	emptyFunc := func(w http.ResponseWriter, r *http.Request) {}
	for i, tc := range []struct {
		http.Handler
	}{
		{http.HandlerFunc(emptyFunc)},
		{Use(middleware1).Use(middleware2).HandlerFunc(emptyFunc)},
		{(&Stack{}).Use(middleware2).HandlerFunc(emptyFunc)},
	} {
		srv := httptest.NewServer(tc.Handler)
		resp, err := http.Get(srv.URL)
		srv.Close()
		resp.Body.Close()
		if err != nil {
			t.Errorf("%d. %v", i, err)
		}
	}
}

func middleware1(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

func middleware2(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}
