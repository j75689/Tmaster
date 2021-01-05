package dbhelper

import (
	"github.com/j75689/Tmaster/pkg/config"
	dbmodel "github.com/j75689/Tmaster/pkg/database/model"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"gorm.io/gorm"
)

const (
	_WorkerName = "DB_Helper"
)

func NewWorker(
	config config.Config,
	db *gorm.DB,
) *DBHelperWorker {
	return &DBHelperWorker{
		config: config,
		db:     db,
	}
}

type DBHelperWorker struct {
	config config.Config
	db     *gorm.DB
}

func (helper *DBHelperWorker) CreateTask(task *dbmodel.Task) error {
	return helper.db.Create(task).Error
}

func (helper *DBHelperWorker) UpdateJob(job *dbmodel.Job) error {
	return helper.db.
		Where("id = ? and ( status = ? or status = ? )", job.ID, model.StatusPending, model.StatusWorking).Updates(job).Error
}
