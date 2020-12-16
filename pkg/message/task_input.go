package message

import "github.com/j75689/Tmaster/pkg/graph/model"

type TaskInput struct {
	Context    Context     `json:"context"`
	InputValue interface{} `json:"input_value"`
	From       string      `json:"from"`
	TaskXID    string      `json:"task_xid"`
	Task       model.Task  `json:"task"`
	Consistent bool        `json:"consistent"`
}
