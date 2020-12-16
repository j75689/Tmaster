package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/endpoint"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

var _ endpoint.Handler = (*GrpcHandler)(nil)

func NewGrpcHandler(logger zerolog.Logger) *GrpcHandler {
	handler := &GrpcHandler{
		logger:   logger,
		connPool: &sync.Map{},
	}
	handler.flushConn()
	return handler
}

type _grpcConn struct {
	conn      *grpc.ClientConn
	flushTime int64
}

type GrpcHandler struct {
	logger   zerolog.Logger
	connPool *sync.Map
}

func (handler *GrpcHandler) flushConn() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			handler.connPool.Range(func(k, v interface{}) bool {
				if client, ok := v.(*_grpcConn); ok {
					if time.Now().Unix() > client.flushTime {
						handler.connPool.Delete(k)
					}
				}
				return true
			})
		}
	}()
}

func (handler *GrpcHandler) getConn(url string, timeout time.Duration) (*grpc.ClientConn, error) {
	var grpcConn *_grpcConn
	if v, ok := handler.connPool.Load(url); ok {
		if client, ok := v.(*_grpcConn); ok {
			client.flushTime = time.Now().Add(time.Hour).Unix()
			handler.connPool.Store(url, client)
			return client.conn, nil
		}
	}

	conn, err := grpc.Dial(url, grpc.WithInsecure(), grpc.WithTimeout(timeout))
	if err != nil {
		return nil, err
	}
	grpcConn = &_grpcConn{
		conn:      conn,
		flushTime: time.Now().Add(time.Hour).Unix(),
	}

	defer func() {
		if grpcConn != nil {
			handler.connPool.Store(url, grpcConn)
		}
	}()
	return conn, err
}

func (handler *GrpcHandler) Do(ctx context.Context, endpointConfig *model.Endpoint) (map[string]string, interface{}, error) {
	handler.logger.Debug().Msg("start grpc endpoint")
	defer handler.logger.Debug().Msg("complete grpc endpoint")
	handler.logger.Debug().Interface("body", endpointConfig).Msgf("endpoint config")

	// open connection
	if endpointConfig.URL == nil {
		return nil, nil, errors.New("no address")
	}

	var (
		url            string
		connectTimeout = time.Second * 10 // default 10 seconds
	)

	if endpointConfig.URL != nil {
		url = *endpointConfig.URL
	}

	if endpointConfig.ConnectTimeout != nil {
		if duration, err := time.ParseDuration(*endpointConfig.ConnectTimeout); err != nil {
			connectTimeout = duration
		}
	}

	// TODO: support ssl connect
	conn, err := handler.getConn(url, connectTimeout)
	if err != nil {
		return nil, nil, err
	}

	format := "json" // json or text
	if endpointConfig.Format != nil {
		format = *endpointConfig.Format
	}
	symbol := ""
	if endpointConfig.Symbol != nil {
		symbol = *endpointConfig.Symbol
	}
	verbose := false
	emitDefaults := false
	includeSeparators := !verbose
	addHeaders := endpointConfig.AddHeaders
	rpcHeaders := endpointConfig.RPCHeaders
	reflHeaders := endpointConfig.ReflHeaders
	data := ""
	if endpointConfig.Body != nil {
		data = *endpointConfig.Body
	}
	in := strings.NewReader(data)

	// TODO: support reflect input struct from config
	// reflect source
	md := grpcurl.MetadataFromHeaders(append(addHeaders, reflHeaders...))
	refCtx := metadata.NewOutgoingContext(ctx, md)
	refClient := grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(conn))
	descSource := grpcurl.DescriptorSourceFromServer(ctx, refClient)

	options := grpcurl.FormatOptions{
		EmitJSONDefaultFields: emitDefaults,
		IncludeTextSeparator:  includeSeparators,
		AllowUnknownFields:    false,
	}
	rf, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.Format(format), descSource, in, options)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to construct request parser and formatter for %q, err = %v", format, err)
	}
	buf := bytes.NewBufferString("")
	h := _NewEventHandler(buf, descSource, formatter, verbose)

	err = grpcurl.InvokeRPC(ctx, descSource, conn, symbol, append(addHeaders, rpcHeaders...), h, rf.Next)
	if err != nil {
		return nil, nil, fmt.Errorf("error invoking method %q, err = %v", symbol, err)
	}

	for _, detected := range endpointConfig.DetectedErrorFromHeader {
		for k, v := range h.Headers() {
			if k == detected.Key && strings.Contains(v, detected.Value) {
				// make error
				err = fmt.Errorf("Header [%s] Contains Error: [%s]", detected.Key, detected.Value)
			}
		}
	}

	// server return error code
	if h.Status.Code() != codes.OK {
		err = fmt.Errorf("ERROR:  Code: %s  Message: %s", h.Status.Code().String(), h.Status.Message())
	}

	var mapData map[string]interface{}
	if merr := json.Unmarshal(buf.Bytes(), &mapData); merr == nil {
		return h.Headers(), mapData, err
	}

	return h.Headers(), buf.String(), err
}

var _ grpcurl.InvocationEventHandler = (*_EventHandler)(nil)

func _NewEventHandler(out io.Writer, descSource grpcurl.DescriptorSource, formatter grpcurl.Formatter, verbose bool) *_EventHandler {
	return &_EventHandler{
		DefaultEventHandler: grpcurl.NewDefaultEventHandler(out, descSource, formatter, verbose),
		headers:             make(map[string]string),
	}
}

type _EventHandler struct {
	*grpcurl.DefaultEventHandler
	headers map[string]string
}

// OnReceiveHeaders is called when response headers have been received.
func (handler *_EventHandler) OnReceiveHeaders(md metadata.MD) {
	for k, v := range md {
		handler.headers[k] = strings.Join(v, ",")
	}
}

func (handler *_EventHandler) Headers() map[string]string {
	return handler.headers
}
