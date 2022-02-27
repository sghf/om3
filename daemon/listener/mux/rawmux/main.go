/*
	Package rawmux provides raw multiplexer from httpmux

	It can be used by raw listeners to Serve accepted connexions
*/
package rawmux

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	clientrequest "opensvc.com/opensvc/core/client/request"
	"opensvc.com/opensvc/daemon/listener/mux/muxresponse"
)

type (
	T struct {
		httpMux http.Handler
		log     zerolog.Logger
		timeOut time.Duration
	}

	ReadWriteCloseSetDeadliner interface {
		io.ReadWriteCloser
		SetDeadline(time.Time) error
	}

	// request struct holds the translated raw request for http mux
	request struct {
		method  string
		path    string
		handler http.HandlerFunc
		body    io.Reader
	}
)

// New function returns an initialised *T that will use mux as http mux
func New(mux http.Handler, log zerolog.Logger, timeout time.Duration) *T {
	return &T{
		httpMux: mux,
		log:     log,
		timeOut: timeout,
	}
}

// Serve function is an adapter to serve raw call from http mux
//
// Serve can be used on raw listeners accepted connexions
//
// 1- raw request will be decoded to create to http request
// 2- http request will be served from http mux ServeHTTP
// 3- Response is sent to w
func (t *T) Serve(w ReadWriteCloseSetDeadliner) {
	defer func() {
		err := w.Close()
		if err != nil {
			t.log.Debug().Err(err).Msg("rawunix.Serve close failure")
			return
		}
	}()
	if err := w.SetDeadline(time.Now().Add(t.timeOut)); err != nil {
		t.log.Error().Err(err).Msg("rawunix.Serve can't set SetDeadline")
	}
	req, err := t.newRequestFrom(w)
	if err != nil {
		t.log.Error().Err(err).Msg("rawunix.Serve can't analyse request")
		return
	}
	resp := muxresponse.NewResponse(w)
	if err := req.do(resp); err != nil {
		t.log.Error().Err(err).Msgf("rawunix.Serve request.do error for %s %s",
			req.method, req.path)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.log.Error().Msgf("rawunix.Serve unexpected status code %d for %s %s",
			resp.StatusCode, req.method, req.path)
		return
	}
	t.log.Info().Msgf("status code is %d", resp.StatusCode)
}

// newRequestFrom functions returns *request from w
func (t *T) newRequestFrom(w io.ReadWriteCloser) (*request, error) {
	var b = make([]byte, 4096)
	_, err := w.Read(b)
	if err != nil {
		t.log.Warn().Err(err).Msg("newRequestFrom read failure")
		return nil, err
	}
	srcRequest := clientrequest.T{}
	b = bytes.TrimRight(b, "\x00")
	if err := json.Unmarshal(b, &srcRequest); err != nil {
		t.log.Warn().Err(err).Msgf("newRequestFrom invalid message: %s", string(b))
		return nil, err
	}
	t.log.Warn().Msgf("newRequestFrom: %s, options: %s", srcRequest, srcRequest.Options)
	matched, ok := actionToPath[srcRequest.Action]
	if !ok {
		msg := "no matched rules for action: " + srcRequest.Action
		return nil, errors.New(msg)
	}
	return &request{
		method:  matched.method,
		path:    matched.path,
		handler: t.httpMux.ServeHTTP,
		body:    bytes.NewReader(b),
	}, nil
}

// do function execute http mux handler on translated request and returns error
func (r *request) do(resp *muxresponse.Response) error {
	body := r.body
	if r.method == "GET" {
		body = nil
	}
	request, err := http.NewRequest(r.method, r.path, body)
	if err != nil {
		return err
	}

	r.handler(resp, request)
	return nil
}
