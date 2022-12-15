// Package daemonapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.0 DO NOT EDIT.
package daemonapi

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (POST /auth/token)
	PostAuthToken(w http.ResponseWriter, r *http.Request)

	// (GET /daemon/events)
	GetDaemonEvents(w http.ResponseWriter, r *http.Request, params GetDaemonEventsParams)

	// (POST /daemon/logs/control)
	PostDaemonLogsControl(w http.ResponseWriter, r *http.Request)

	// (GET /daemon/running)
	GetDaemonRunning(w http.ResponseWriter, r *http.Request)

	// (GET /daemon/status)
	GetDaemonStatus(w http.ResponseWriter, r *http.Request, params GetDaemonStatusParams)

	// (POST /daemon/stop)
	PostDaemonStop(w http.ResponseWriter, r *http.Request)

	// (POST /daemon/sub/action)
	PostDaemonSubAction(w http.ResponseWriter, r *http.Request)

	// (POST /node/clear)
	PostNodeClear(w http.ResponseWriter, r *http.Request)

	// (POST /node/monitor)
	PostNodeMonitor(w http.ResponseWriter, r *http.Request)

	// (GET /nodes/info)
	GetNodesInfo(w http.ResponseWriter, r *http.Request)

	// (POST /object/abort)
	PostObjectAbort(w http.ResponseWriter, r *http.Request)

	// (POST /object/clear)
	PostObjectClear(w http.ResponseWriter, r *http.Request)

	// (GET /object/config)
	GetObjectConfig(w http.ResponseWriter, r *http.Request, params GetObjectConfigParams)

	// (GET /object/file)
	GetObjectFile(w http.ResponseWriter, r *http.Request, params GetObjectFileParams)

	// (POST /object/monitor)
	PostObjectMonitor(w http.ResponseWriter, r *http.Request)

	// (GET /object/selector)
	GetObjectSelector(w http.ResponseWriter, r *http.Request, params GetObjectSelectorParams)

	// (POST /object/status)
	PostObjectStatus(w http.ResponseWriter, r *http.Request)

	// (GET /public/openapi)
	GetSwagger(w http.ResponseWriter, r *http.Request)

	// (GET /relay/message)
	GetRelayMessage(w http.ResponseWriter, r *http.Request, params GetRelayMessageParams)

	// (POST /relay/message)
	PostRelayMessage(w http.ResponseWriter, r *http.Request)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// PostAuthToken operation middleware
func (siw *ServerInterfaceWrapper) PostAuthToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostAuthToken(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetDaemonEvents operation middleware
func (siw *ServerInterfaceWrapper) GetDaemonEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetDaemonEventsParams

	// ------------- Optional query parameter "duration" -------------
	if paramValue := r.URL.Query().Get("duration"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "duration", r.URL.Query(), &params.Duration)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "duration", Err: err})
		return
	}

	// ------------- Optional query parameter "limit" -------------
	if paramValue := r.URL.Query().Get("limit"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "limit", r.URL.Query(), &params.Limit)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "limit", Err: err})
		return
	}

	// ------------- Optional query parameter "filter" -------------
	if paramValue := r.URL.Query().Get("filter"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "filter", r.URL.Query(), &params.Filter)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "filter", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetDaemonEvents(w, r, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PostDaemonLogsControl operation middleware
func (siw *ServerInterfaceWrapper) PostDaemonLogsControl(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostDaemonLogsControl(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetDaemonRunning operation middleware
func (siw *ServerInterfaceWrapper) GetDaemonRunning(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetDaemonRunning(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetDaemonStatus operation middleware
func (siw *ServerInterfaceWrapper) GetDaemonStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetDaemonStatusParams

	// ------------- Optional query parameter "namespace" -------------
	if paramValue := r.URL.Query().Get("namespace"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "namespace", r.URL.Query(), &params.Namespace)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "namespace", Err: err})
		return
	}

	// ------------- Optional query parameter "relatives" -------------
	if paramValue := r.URL.Query().Get("relatives"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "relatives", r.URL.Query(), &params.Relatives)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "relatives", Err: err})
		return
	}

	// ------------- Optional query parameter "selector" -------------
	if paramValue := r.URL.Query().Get("selector"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "selector", r.URL.Query(), &params.Selector)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "selector", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetDaemonStatus(w, r, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PostDaemonStop operation middleware
func (siw *ServerInterfaceWrapper) PostDaemonStop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostDaemonStop(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PostDaemonSubAction operation middleware
func (siw *ServerInterfaceWrapper) PostDaemonSubAction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostDaemonSubAction(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PostNodeClear operation middleware
func (siw *ServerInterfaceWrapper) PostNodeClear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostNodeClear(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PostNodeMonitor operation middleware
func (siw *ServerInterfaceWrapper) PostNodeMonitor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostNodeMonitor(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetNodesInfo operation middleware
func (siw *ServerInterfaceWrapper) GetNodesInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetNodesInfo(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PostObjectAbort operation middleware
func (siw *ServerInterfaceWrapper) PostObjectAbort(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostObjectAbort(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PostObjectClear operation middleware
func (siw *ServerInterfaceWrapper) PostObjectClear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostObjectClear(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetObjectConfig operation middleware
func (siw *ServerInterfaceWrapper) GetObjectConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetObjectConfigParams

	// ------------- Required query parameter "path" -------------
	if paramValue := r.URL.Query().Get("path"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "path"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "path", r.URL.Query(), &params.Path)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "path", Err: err})
		return
	}

	// ------------- Optional query parameter "evaluate" -------------
	if paramValue := r.URL.Query().Get("evaluate"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "evaluate", r.URL.Query(), &params.Evaluate)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "evaluate", Err: err})
		return
	}

	// ------------- Optional query parameter "impersonate" -------------
	if paramValue := r.URL.Query().Get("impersonate"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "impersonate", r.URL.Query(), &params.Impersonate)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "impersonate", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetObjectConfig(w, r, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetObjectFile operation middleware
func (siw *ServerInterfaceWrapper) GetObjectFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetObjectFileParams

	// ------------- Required query parameter "path" -------------
	if paramValue := r.URL.Query().Get("path"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "path"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "path", r.URL.Query(), &params.Path)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "path", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetObjectFile(w, r, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PostObjectMonitor operation middleware
func (siw *ServerInterfaceWrapper) PostObjectMonitor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostObjectMonitor(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetObjectSelector operation middleware
func (siw *ServerInterfaceWrapper) GetObjectSelector(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetObjectSelectorParams

	// ------------- Required query parameter "selector" -------------
	if paramValue := r.URL.Query().Get("selector"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "selector"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "selector", r.URL.Query(), &params.Selector)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "selector", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetObjectSelector(w, r, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PostObjectStatus operation middleware
func (siw *ServerInterfaceWrapper) PostObjectStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostObjectStatus(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetSwagger operation middleware
func (siw *ServerInterfaceWrapper) GetSwagger(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetSwagger(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetRelayMessage operation middleware
func (siw *ServerInterfaceWrapper) GetRelayMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetRelayMessageParams

	// ------------- Optional query parameter "nodename" -------------
	if paramValue := r.URL.Query().Get("nodename"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "nodename", r.URL.Query(), &params.Nodename)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "nodename", Err: err})
		return
	}

	// ------------- Optional query parameter "cluster_id" -------------
	if paramValue := r.URL.Query().Get("cluster_id"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "cluster_id", r.URL.Query(), &params.ClusterId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "cluster_id", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetRelayMessage(w, r, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PostRelayMessage operation middleware
func (siw *ServerInterfaceWrapper) PostRelayMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = context.WithValue(ctx, BasicAuthScopes, []string{""})

	ctx = context.WithValue(ctx, BearerAuthScopes, []string{""})

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostRelayMessage(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/auth/token", wrapper.PostAuthToken)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/daemon/events", wrapper.GetDaemonEvents)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/daemon/logs/control", wrapper.PostDaemonLogsControl)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/daemon/running", wrapper.GetDaemonRunning)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/daemon/status", wrapper.GetDaemonStatus)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/daemon/stop", wrapper.PostDaemonStop)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/daemon/sub/action", wrapper.PostDaemonSubAction)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/node/clear", wrapper.PostNodeClear)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/node/monitor", wrapper.PostNodeMonitor)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/nodes/info", wrapper.GetNodesInfo)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/object/abort", wrapper.PostObjectAbort)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/object/clear", wrapper.PostObjectClear)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/object/config", wrapper.GetObjectConfig)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/object/file", wrapper.GetObjectFile)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/object/monitor", wrapper.PostObjectMonitor)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/object/selector", wrapper.GetObjectSelector)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/object/status", wrapper.PostObjectStatus)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/public/openapi", wrapper.GetSwagger)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/relay/message", wrapper.GetRelayMessage)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/relay/message", wrapper.PostRelayMessage)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
