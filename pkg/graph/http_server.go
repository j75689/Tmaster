package graph

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/graph/generated"
	"github.com/j75689/Tmaster/pkg/opentracer"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog"
)

func NewHttpServer(
	config config.Config,
	graphqlConfig generated.Config,
	openTracer opentracer.OpenTracer,
	logger zerolog.Logger,
) *HttpServer {
	if openTracer != nil {
		opentracing.SetGlobalTracer(openTracer.Tracer())
	}
	return &HttpServer{
		httpServer:    &http.Server{},
		config:        config,
		graphqlConfig: graphqlConfig,
		logger:        logger,
	}

}

type HttpServer struct {
	httpServer    *http.Server
	config        config.Config
	graphqlConfig generated.Config
	logger        zerolog.Logger
}

func (server *HttpServer) Start() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(logger.SetLogger(
		logger.WithUTC(true),
		logger.WithLogger(func(c *gin.Context, w io.Writer, latency time.Duration) zerolog.Logger {
			return server.logger.With().
				Str("path", c.Request.URL.Path).
				Str("latency", latency.String()).
				Logger()
		})))
	if !server.config.HTTP.Graphql.Playground.Disable {
		router.Any(server.config.HTTP.Graphql.Playground.Path, server.PlaygroundHandler(server.config.HTTP.Graphql.Playground.Title, server.config.HTTP.Graphql.Endpoint))
	}
	router.Any(server.config.HTTP.Graphql.Endpoint, server.ActionHandler(server.graphqlConfig))
	server.logger.Info().Msgf("listen on port :%d", server.config.HTTP.Port)
	server.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", server.config.HTTP.Port),
		Handler: router,
	}
	return server.httpServer.ListenAndServe()
}

func (server *HttpServer) Shutdown() error {
	return server.httpServer.Shutdown(context.Background())
}
