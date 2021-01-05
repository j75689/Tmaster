package model

import (
	"time"

	"github.com/j75689/Tmaster/pkg/graph/model"
)

type Task struct {
	ID           int64
	Name         string `gorm:"type:varchar(1024)"`
	Created      time.Time
	Updated      time.Time
	TaskHistory  *model.TaskHistory `gorm:"embedded"`
	JobID        int64              `gorm:"index"`
	ErrorCode    model.ErrorCode    `gorm:"type:varchar(255)"`
	ErrorMessage string             `gorm:"type:text"`
}
