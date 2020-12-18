package http

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/j75689/Tmaster/pkg/endpoint"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"github.com/rs/zerolog"
)

var _ endpoint.Handler = (*HttpHandler)(nil)

func NewHttpHandler(logger zerolog.Logger) *HttpHandler {
	return &HttpHandler{
		logger: logger,
	}
}

type HttpHandler struct {
	logger zerolog.Logger
}

func (handler *HttpHandler) Do(ctx context.Context, endpointConfig *model.Endpoint) (map[string]string, interface{}, error) {
	handler.logger.Debug().Msg("start http endpoint")
	defer handler.logger.Debug().Msg("complete http endpoint")
	handler.logger.Debug().Interface("body", endpointConfig).Msgf("endpoint config")

	var (
		client = &http.Client{}
		proxy  func(*http.Request) (*url.URL, error)
	)
	if endpointConfig.Proxy != nil {
		proxyURL, err := url.Parse(*endpointConfig.Proxy)
		if err != nil {
			return nil, nil, err
		}
		proxy = http.ProxyURL(proxyURL)
	} else {
		proxy = http.ProxyFromEnvironment
	}

	insecure := false
	if endpointConfig.Insecure != nil {
		insecure = *endpointConfig.Insecure
	}

	client.Transport = &http.Transport{
		Proxy: proxy,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	method := ""
	if endpointConfig.Method != nil {
		method = endpointConfig.Method.String()
	}

	url := ""
	if endpointConfig.URL != nil {
		url = *endpointConfig.URL
	}
	requestBody := ""
	if endpointConfig.Body != nil {
		requestBody = *endpointConfig.Body
	}
	req, err := http.NewRequest(method, url, strings.NewReader(requestBody))
	if err != nil {
		return nil, nil, err
	}
	req = req.WithContext(ctx)
	for _, item := range endpointConfig.Headers {
		req.Header.Set(item.Key, item.Value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var (
		body   []byte
		header = make(map[string]string)
		// Check that the server actually sent compressed data
		reader io.ReadCloser
	)
	switch strings.ToUpper(resp.Header.Get("Content-Encoding")) {
	case "GZIP":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	for k, v := range resp.Header {
		header[k] = strings.Join(v, ",")
	}

	body, err = ioutil.ReadAll(reader)
	if err != nil {
		return header, nil, err
	}

	for _, detected := range endpointConfig.DetectedErrorFromHeader {
		for k, v := range header {
			if k == detected.Key && strings.Contains(v, detected.Value) {
				// make error
				err = fmt.Errorf("Header [%s] Contains Error: [%s]", detected.Key, detected.Value)
			}
		}
	}

	var mapData map[string]interface{}
	if merr := json.Unmarshal(body, &mapData); merr == nil {
		return header, mapData, err
	}

	return header, string(body), err
}
