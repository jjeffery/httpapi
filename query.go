// Copyright 2016 John Jeffery <john@jeffery.id.au>. All rights reserved.

package httpapi

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jjeffery/errkind"
	"github.com/jjeffery/stringset"
	"github.com/spkg/local"
)

// Values provides convenient methods for extracting arguments from the query string.
type Values struct {
	values        url.Values
	invalidParams stringset.Set
}

// Query returns values from the query string part of the request URL.
func Query(r *http.Request) *Values {
	return &Values{
		values:        r.URL.Query(),
		invalidParams: stringset.New(),
	}
}

// Err returns nil if no errors have been encountered, otherwise it
// returns a bad request error that lists the parameter(s) that are
// not in the correct format.
func (v *Values) Err() error {
	if v.invalidParams.Len() == 0 {
		return nil
	}
	// We want the client to know which parameters, so we have to format them
	// in the error message.
	msg := fmt.Sprintf("invalid value(s) in query string: %s", strings.Join(v.invalidParams.Values(), ","))
	err := errkind.BadRequest(msg)
	return err
}

// validate runs a validation function over all parameters with the
// specified names. Returns the first error encountered, or nil if no errors.
func (v *Values) validate(names []string, validator func(string)) {
	m := map[string][]string(v.values)
	for _, name := range names {
		vals, ok := m[name]
		if !ok {
			continue
		}
		for _, val := range vals {
			validator(val)
		}
	}
}

// LookupInt returns an integer, with an indication of whether the
// query value was present.
func (v *Values) LookupInt(name string) (n int, ok bool) {
	return v.parseInt(name)
}

// GetInt returns an int. Returns 0 if the query value is not
// present in the query.
func (v *Values) GetInt(name string) int {
	n, _ := v.parseInt(name)
	return n
}

// LookupTime returns a time. The time should be in RFC3339 format.
func (v *Values) LookupTime(name string) (t time.Time, ok bool) {
	return v.parseTime(name)
}

// GetTime returns a time. The time should be in RFC3339 format.
// Returns zero if the time value if not present in the query.
func (v *Values) GetTime(name string) time.Time {
	t, _ := v.parseTime(name)
	return t
}

// LookupDate returns a date. The date should be in ISO8601 format.
func (v *Values) LookupDate(name string) (d local.Date, ok bool) {
	return v.parseDate(name)
}

// GetDate returns a date. The date should be in ISO8601 format.
// Returns zero if the date value if not present in the query.
func (v *Values) GetDate(name string) local.Date {
	d, _ := v.parseDate(name)
	return d
}

// LookupBool returns a bool, with an indication of whether the
// query value was present in the query.
func (v *Values) LookupBool(name string) (b bool, ok bool) {
	return v.parseBool(name)
}

// GetBool returns a bool. Returns false if the query value is not present
// in the query.
func (v *Values) GetBool(name string) bool {
	b, _ := v.parseBool(name)
	return b
}

// LookupString returns a string, with an indication of whether the
// query value was present in the query.
func (v *Values) LookupString(name string) (s string, ok bool) {
	if v.exists(name) {
		return v.values.Get(name), true
	}
	return "", false
}

// GetString returns a string. Returns "" if the query value is not
// present in the query.
func (v *Values) GetString(name string) string {
	if v.exists(name) {
		return v.values.Get(name)
	}
	return ""
}

func (v *Values) exists(name string) bool {
	_, ok := v.values[name]
	return ok
}

func (v *Values) parseTime(name string) (time.Time, bool) {
	if !v.exists(name) {
		return time.Time{}, false
	}
	s := v.values.Get(name)
	s = strings.TrimSpace(s)
	if s == "" || s == "undefined" || s == "null" {
		return time.Time{}, false
	}

	var t time.Time
	var err error

	if t, err = time.Parse(time.RFC3339Nano, s); err != nil {
		if t, err = time.Parse(time.RFC3339, s); err != nil {
			v.invalidParams.Add(name)
			return time.Time{}, false
		}
	}
	return t, true
}

func (v *Values) parseDate(name string) (local.Date, bool) {
	if !v.exists(name) {
		return local.Date{}, false
	}
	s := v.values.Get(name)
	s = strings.TrimSpace(s)
	if s == "" || s == "undefined" || s == "null" {
		return local.Date{}, false
	}

	var d local.Date
	var err error

	if d, err = local.DateParse(s); err != nil {
		v.invalidParams.Add(name)
		return local.Date{}, false
	}
	return d, true
}

func (v *Values) parseInt(name string) (int, bool) {
	if !v.exists(name) {
		return 0, false
	}
	s := v.values.Get(name)
	var n int
	var err error
	if n, err = strconv.Atoi(s); err != nil {
		v.invalidParams.Add(name)
		return 0, false
	}
	return n, true
}

func (v *Values) parseBool(name string) (bool, bool) {
	if !v.exists(name) {
		return false, false
	}
	s := strings.ToLower(v.values.Get(name))
	switch s {
	case "1", "true", "yes", "t":
		return true, true
	case "0", "false", "no", "f":
		return false, true
	}
	v.invalidParams.Add(name)
	return false, false
}
