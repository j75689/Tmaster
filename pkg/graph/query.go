package graph

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gin-gonic/gin"
	"github.com/j75689/Tmaster/pkg/graph/generated"
)

// ActionHandler reutnrs a gin.handlerFunc that handling graphql action
func (server *HttpServer) ActionHandler(graphqlConfig generated.Config) gin.HandlerFunc {
	h := handler.NewDefaultServer(generated.NewExecutableSchema(graphqlConfig))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
