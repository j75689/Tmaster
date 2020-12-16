package graph

import (
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

// PlaygroundHandler returns gin.HaandlerFunc that handling graphql playground
func (server *HttpServer) PlaygroundHandler(title string, endpoint string) gin.HandlerFunc {
	h := playground.Handler(title, endpoint)

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
