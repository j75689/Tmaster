package wireset

import (
	"github.com/google/wire"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/graph/generated"
	"github.com/j75689/Tmaster/pkg/graph/resolver"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/j75689/Tmaster/pkg/opentracer"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var GraphqlSchemaSet = wire.NewSet(
	InitializeGraphqlSchema,
)

func InitializeGraphqlSchema(config config.Config, mq mq.MQ, db *gorm.DB, tracer *opentracer.ServiceTracer, logger zerolog.Logger) generated.Config {
	return generated.Config{
		Resolvers: resolver.NewResolver(config, mq, db, tracer, logger),
	}
}
