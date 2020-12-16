package model

import (
	"time"

	"github.com/j75689/Tmaster/pkg/graph/model"
)

type Task struct {
	ID                 int64
	Name               string
	Created            time.Time `xorm:"created"`
	Updated            time.Time `xorm:"updated"`
	*model.TaskHistory `xorm:"extends"`
	JobID              int64           `xorm:"index"`
	ErrorCode          model.ErrorCode `xorm:"error_code"`
	ErrorMessage       string          `xorm:"'error_message' text"`
}
