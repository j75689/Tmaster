package model

import (
	"time"

	"github.com/j75689/Tmaster/pkg/graph/model"
)

type Job struct {
	ID               int64
	ParentID         *int64
	Created          time.Time `xorm:"created"`
	*model.JobStatus `xorm:"extends"`
}
