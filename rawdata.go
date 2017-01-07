// Copyright 2016 John Jeffery <john@jeffery.id.au>. All rights reserved.

package httpapi

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/jjeffery/errkind"
	"github.com/jjeffery/errors"
)

// maxRequestLen is the max size we are prepared to read from a HTTP client.
// Anything this size or larger gets discarded.
var maxRequestLen = 1024 * 1024 * 16

// Content encodings
const (
	ceIdentity = "identity"
	ceDeflate  = "deflate"
	ceGzip     = "gzip"
)

// rawData represents a data BLOB that can be read from or written to
// persistent storage, or a HTTP client.
type rawData struct {
	ContentType        string
	ContentEncoding    string
	Content            []byte
	UncompressedLength int
}

// IsCompressed returns whether the content is compressed.
func (data *rawData) IsCompressed() bool {
	if data.ContentEncoding == "" {
		data.ContentEncoding = ceIdentity
	}
	return data.ContentEncoding != ceIdentity
}

// ReadRequest reads the data from the request into the raw.Data.
func (data *rawData) ReadRequest(r *http.Request) error {
	if cl := r.Header.Get("Content-Length"); cl != "" {
		v, err := strconv.ParseInt(cl, 10, 64)
		if err != nil || v < 0 {
			return errkind.BadRequest("invalid content-length")
		}

		if v >= int64(maxRequestLen) {
			return errkind.Public("payload too large", http.StatusRequestEntityTooLarge)
		}

		buf := make([]byte, v)

		_, err = io.ReadFull(r.Body, buf)
		if err != nil {
			return errkind.BadRequest("cannot read full content")
		}
		data.Content = buf
	} else {
		reader := io.LimitReader(r.Body, int64(maxRequestLen))
		content, err := ioutil.ReadAll(reader)
		if err != nil {
			return errkind.BadRequest("cannot read all content")
		}
		if len(content) >= maxRequestLen {
			return errkind.Public("payload too large", http.StatusRequestEntityTooLarge)
		}
		data.Content = content
	}

	// The HTTP specification does not mention Content-Encoding for
	// requests, but sometimes it is handy to allow the client to do so.
	if ce := r.Header.Get("Content-Encoding"); ce != "" {
		data.ContentEncoding = ce
		data.UncompressedLength = 0 // not known
	} else {
		data.UncompressedLength = len(data.Content)
		data.ContentEncoding = ceIdentity
	}

	data.ContentType = r.Header.Get("Content-Type")
	if data.ContentType == "" {
		data.ContentType = "application/octet-stream"
	}
	return nil
}

// WriteResponse writes the contents to the client as a response.
func (data *rawData) WriteResponse(w http.ResponseWriter) error {
	if len(data.Content) == 0 {
		w.Header().Set("Content-Length", "0")
		w.Header().Del("Content-Type")
		w.Header().Del("Content-Encoding")
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	if data.IsCompressed() {
		w.Header().Set("Content-Encoding", data.ContentEncoding)
	} else {
		w.Header().Del("Content-Encoding")
	}
	w.Header().Set("Content-Type", data.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data.Content)))
	_, err := w.Write(data.Content)
	if err != nil {
		return errors.Wrap(err, "cannot write response")
	}
	return nil
}

func (data *rawData) Decompress() error {
	if !data.IsCompressed() {
		return nil
	}
	input := bytes.NewBuffer(data.Content)
	var reader io.Reader
	if data.ContentEncoding == ceDeflate {
		reader = flate.NewReader(input)
	} else if data.ContentEncoding == ceGzip {
		var err error
		if reader, err = gzip.NewReader(input); err != nil {
			return err
		}
	} else {
		return errors.New("unknown content-encoding").
			With("content-encoding", data.ContentEncoding)
	}
	writer := bytes.Buffer{}
	_, err := io.Copy(&writer, reader)
	if err != nil {
		return err
	}
	data.Content = writer.Bytes()
	data.ContentEncoding = ""
	data.UncompressedLength = len(data.Content)
	return nil
}

func (data *rawData) CompressResponse(r *http.Request) error {
	// additional overhead in compressed response
	const overhead = 24 // len("Content-Encoding: gzip\r\n")

	if data.IsCompressed() || len(data.Content) < overhead*4 {
		// already compressed, or not worth compressing
		// because data is nil or too short
		return nil
	}

	// TODO(jpj): this is a fairly naive handling of the Accept-Encoding
	// header. In particular it does not handle gzip;q=0, which is
	// a valid way of saying that gzip is not acceptable.
	if ae := r.Header.Get("Accept-Encoding"); !strings.Contains(ae, ceGzip) {
		return nil
	}

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	n, err := w.Write(data.Content)
	if err != nil {
		return err
	}
	if n != len(data.Content) {
		return errors.New("cannot compress")
	}
	err = w.Close()
	if err != nil {
		return err
	}
	compressedBytes := buf.Bytes()

	if len(compressedBytes)+overhead < len(data.Content) {
		data.UncompressedLength = len(data.Content)
		data.Content = compressedBytes
		data.ContentEncoding = ceGzip
	}

	return nil
}

func (data *rawData) UnmarshalTo(v interface{}) error {
	err := data.Decompress()
	if err != nil {
		return errkind.BadRequest("cannot decompress payload")
	}
	err = json.Unmarshal(data.Content, v)
	if err != nil {
		return errkind.BadRequest("invalid JSON payload")
	}
	return nil
}

func (data *rawData) MarshalFrom(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data.Content = b
	data.ContentType = "application/json"
	data.ContentEncoding = ""
	data.UncompressedLength = len(b)
	return nil
}
