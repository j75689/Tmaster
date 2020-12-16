package message

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/j75689/Tmaster/pkg/graph/model"
)

type Context struct {
	context.Context
	TraceCarrier opentracing.TextMapCarrier `json:"trace_carrier"`
	Execution    Execution                  `json:"execution"`
	State        State                      `json:"state"`
	Job          Job                        `json:"job"`
	Tasks        map[string]*model.Task     `json:"tasks"`
}

type Execution struct {
	ID                string           `json:"id"`
	StartTime         time.Time        `json:"start_time"`
	EndTime           time.Time        `json:"end_time"`
	Cause             model.Cause      `json:"cause"`
	CauseError        *string          `json:"cause_error,omitempty"`
	CauseErrorCode    *model.ErrorCode `json:"cause_error_code,omitempty"`
	MaxConsistentNums int              `json:"max_consistent_nums"`
	ConsistentNums    int              `json:"consistent_nums"`
}

type State struct {
	EnteredTime time.Time    `json:"entered_time"`
	Name        string       `json:"name"`
	RetryCount  int          `json:"retry_count"`
	Status      model.Status `json:"status"`
}

type Job struct {
	ID    int64  `json:"id"`
	JobID string `jsonb:"job_id"`
}
