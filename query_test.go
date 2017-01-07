// Copyright 2016 John Jeffery <john@jeffery.id.au>. All rights reserved.

package httpapi

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/spkg/local"
)

func TestQuery(t *testing.T) {
	tests := []struct {
		url     string
		bools   map[string]bool
		ints    map[string]int
		times   map[string]time.Time
		strings map[string]string
		dates   map[string]local.Date
	}{
		{
			url: "https://xyris.io/?bool=true&int=12&time=2020-01-02T13:14:15Z&string=string!&date=2099-12-31",
			bools: map[string]bool{
				"bool": true,
			},
			ints: map[string]int{
				"int": 12,
			},
			strings: map[string]string{
				"string": "string!",
			},
			times: map[string]time.Time{
				"time": time.Date(2020, 1, 2, 13, 14, 15, 0, time.UTC),
			},
			dates: map[string]local.Date{
				"date": local.DateFor(2099, 12, 31),
			},
		},
		{
			url: "https://xyris.io/?bool=1&bool=0&int=1&int=2",
			bools: map[string]bool{
				"bool": true,
			},
			ints: map[string]int{
				"int": 1,
			},
		},
		{
			url: "https://xyris.io/?b1=t&b2=f&b3=TRUE&b4=FALSE",
			bools: map[string]bool{
				"b1": true,
				"b2": false,
				"b3": true,
				"b4": false,
			},
		},
		{
			url: "https://xyris.io/?t1=2020-01-02T13:14:15.123456789Z",
			times: map[string]time.Time{
				"t1": time.Date(2020, 1, 2, 13, 14, 15, 123456789, time.UTC),
			},
		},
	}

	for i, tt := range tests {
		rURL, err := url.Parse(tt.url)
		if err != nil {
			t.Errorf("%d: cannot parse url %s: %v", i, tt.url, err)
			continue
		}
		r := &http.Request{
			URL: rURL,
		}

		query := Query(r)
		for name, want := range tt.bools {
			got, ok := query.LookupBool(name)
			if !ok {
				t.Errorf("%d: expected %q, found none", i, name)
			}
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			got = query.GetBool(name)
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			name = name + "_not_present"
			want = false
			got, ok = query.LookupBool(name)
			if ok {
				t.Errorf("%d: expected no %q, found %v", i, name, got)
			}
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			got = query.GetBool(name)
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
		}
		for name, want := range tt.ints {
			got, ok := query.LookupInt(name)
			if !ok {
				t.Errorf("%d: expected %q, found none", i, name)
			}
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			got = query.GetInt(name)
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			name = name + "_not_present"
			want = 0
			got, ok = query.LookupInt(name)
			if ok {
				t.Errorf("%d: expected no %q, found %v", i, name, got)
			}
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			got = query.GetInt(name)
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
		}
		for name, want := range tt.times {
			got, ok := query.LookupTime(name)
			if !ok {
				t.Errorf("%d: expected %q, found none", i, name)
			}
			if !got.Equal(want) {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			got = query.GetTime(name)
			if !got.Equal(want) {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			name = name + "_not_present"
			want = time.Time{}
			got, ok = query.LookupTime(name)
			if ok {
				t.Errorf("%d: expected no %q, found %v", i, name, got)
			}
			if !got.Equal(want) {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			got = query.GetTime(name)
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
		}
		for name, want := range tt.dates {
			got, ok := query.LookupDate(name)
			if !ok {
				t.Errorf("%d: expected %q, found none", i, name)
			}
			if !got.Equal(want) {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			got = query.GetDate(name)
			if !got.Equal(want) {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			name = name + "_not_present"
			want = local.Date{}
			got, ok = query.LookupDate(name)
			if ok {
				t.Errorf("%d: expected no %q, found %v", i, name, got)
			}
			if !got.Equal(want) {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			got = query.GetDate(name)
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
		}
		for name, want := range tt.strings {
			got, ok := query.LookupString(name)
			if !ok {
				t.Errorf("%d: expected %q, found none", i, name)
			}
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			got = query.GetString(name)
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			name = name + "_not_present"
			want = ""
			got, ok = query.LookupString(name)
			if ok {
				t.Errorf("%d: expected no %q, found %v", i, name, got)
			}
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
			got = query.GetString(name)
			if got != want {
				t.Errorf("%d: %q: want %v, got %v", i, name, want, got)
			}
		}
	}
}
