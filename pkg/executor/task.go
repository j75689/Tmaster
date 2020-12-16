package executor

import (
	"github.com/j75689/Tmaster/pkg/errors"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"github.com/j75689/Tmaster/pkg/message"
)

// Executor is an interface for task executor
type Executor interface {
	Execute(message.Context, interface{}, *model.Task) (interface{}, *model.Task, errors.Error)
}
