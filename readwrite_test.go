package httpapi

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/jjeffery/errkind"
)

func readCloserFromString(s string) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader([]byte(s)))
}

type errorReadCloser struct{}

func (e errorReadCloser) Read(data []byte) (int, error) {
	return 0, errors.New("read failed")
}

func (e errorReadCloser) Close() error {
	return errors.New("close failed")
}

type infiniteReadCloser struct{}

func (e infiniteReadCloser) Read(data []byte) (int, error) {
	for i := 0; i < len(data); i++ {
		data[i] = ' '
	}
	return len(data), nil
}

func (e infiniteReadCloser) Close() error {
	return nil
}

func TestReadRequest(t *testing.T) {
	type Payload struct {
		String string
		Int    int
	}
	tests := []struct {
		header     http.Header
		body       io.ReadCloser
		want       Payload
		wantStatus int
	}{
		{
			header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			body: readCloserFromString(`{"String":"S","Int":99}`),
			want: Payload{String: "S", Int: 99},
		},
		{
			header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			body:       readCloserFromString(`{"String":"S","Int":`),
			wantStatus: http.StatusBadRequest,
		},
		{
			header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			body:       errorReadCloser{},
			wantStatus: http.StatusBadRequest,
		},
		{
			header: http.Header{
				"Content-Type":   []string{"application/json"},
				"Content-Length": []string{"9999999999"},
			},
			body:       errorReadCloser{},
			wantStatus: http.StatusRequestEntityTooLarge,
		},
		{
			header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			body:       infiniteReadCloser{},
			wantStatus: http.StatusRequestEntityTooLarge,
		},
	}
	for i, tt := range tests {
		r := http.Request{
			Header: tt.header,
			Body:   tt.body,
		}
		var got Payload
		err := ReadRequest(&r, &got)
		if err != nil {
			if tt.wantStatus == 0 {
				t.Errorf("%d: want no error got %v", i, err)
			}
			if status := errkind.Status(err); status != tt.wantStatus {
				t.Errorf("%d: want status=%d, got %d", i, tt.wantStatus, status)
			}
			continue
		}
		if got != tt.want {
			t.Errorf("%d: want %v got %v", i, tt.want, got)
		}
	}
}

func TestWriteResponse(t *testing.T) {

}
