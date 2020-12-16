package message

import "github.com/j75689/Tmaster/pkg/graph/model"

type InitJob struct {
	Context     Context `json:"context"`
	JobID       string  `json:"job_id"`
	ParentJobID *string `json:"parent_job_id,omitempty"`
	ParentID    *int64  `json:"parent_id,omitempty"`
	model.Job   `json:",inline"`
}
