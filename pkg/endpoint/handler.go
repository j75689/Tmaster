package endpoint

import (
	"context"

	"github.com/j75689/Tmaster/pkg/graph/model"
)

// Handler is an interface for endpoint handler
type Handler interface {
	Do(context.Context, *model.Endpoint) (header map[string]string, body interface{}, err error)
}
