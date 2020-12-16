package wireset

import (
	"github.com/google/wire"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/graph"
	"github.com/j75689/Tmaster/pkg/graph/generated"
	"github.com/j75689/Tmaster/pkg/opentracer"
	"github.com/rs/zerolog"
)

var HttpSet = wire.NewSet(
	InitializeHttpServer,
)

func InitializeHttpServer(
	config config.Config,
	graphqlConfig generated.Config,
	openTracer opentracer.OpenTracer,
	logger zerolog.Logger,
) *graph.HttpServer {
	return graph.NewHttpServer(config, graphqlConfig, openTracer, logger)
}
