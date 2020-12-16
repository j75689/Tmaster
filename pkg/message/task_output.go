package message

import "github.com/j75689/Tmaster/pkg/graph/model"

type TaskOutput struct {
	Context     Context          `json:"context"`
	ErrorCode   *model.ErrorCode `json:"error_code,omitempty"`
	Error       *string          `json:"error,omitempty"`
	InputValue  interface{}      `json:"input_value"`
	OutputValue interface{}      `json:"output_value"`
	From        string           `json:"from"`
	TaskXID     string           `json:"task_xid"`
	Task        model.Task       `json:"task"`
}
