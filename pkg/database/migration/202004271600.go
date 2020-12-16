package migration

import (
	"github.com/j75689/Tmaster/pkg/database/model"
	"xorm.io/xorm"
	"xorm.io/xorm/migrate"
)

var v202004271600 = &migrate.Migration{
	ID: "202004271600",
	Migrate: func(tx *xorm.Engine) error {
		if err := tx.Sync2(&model.Job{}); err != nil {
			return err
		}
		if err := tx.Sync2(&model.Task{}); err != nil {
			return err
		}
		return nil
	},
	Rollback: func(tx *xorm.Engine) error {
		if err := tx.DropTables(&model.Job{}); err != nil {
			return err
		}
		if err := tx.DropTables(&model.Task{}); err != nil {
			return err
		}
		return nil
	},
}
