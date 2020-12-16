package dbhelper

import (
	"github.com/j75689/Tmaster/pkg/config"
	dbmodel "github.com/j75689/Tmaster/pkg/database/model"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"xorm.io/builder"
	"xorm.io/xorm"
)

const (
	_WorkerName = "DB_Helper"
)

func NewWorker(
	config config.Config,
	db *xorm.Engine,
) *DBHelperWorker {
	return &DBHelperWorker{
		config: config,
		db:     db,
	}
}

type DBHelperWorker struct {
	config config.Config
	db     *xorm.Engine
}

func (helper *DBHelperWorker) CreateTask(task *dbmodel.Task) error {
	_, err := helper.db.InsertOne(task)
	return err
}

func (helper *DBHelperWorker) UpdateJob(job *dbmodel.Job) error {
	_, err := helper.db.ID(job.ID).Where(builder.Or(
		builder.Eq{"status": model.StatusPending},
		builder.Eq{"status": model.StatusWorking},
	)).Update(job)
	return err
}
