package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/j75689/Tmaster/pkg/database/model"
	"gorm.io/gorm"
)

var v202004271600 = &gormigrate.Migration{
	ID: "202004271600",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Job{}); err != nil {
			return err
		}
		if err := tx.AutoMigrate(&model.Task{}); err != nil {
			return err
		}
		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		if err := tx.Migrator().DropTable(&model.Job{}); err != nil {
			return err
		}
		if err := tx.Migrator().DropTable(&model.Task{}); err != nil {
			return err
		}
		return nil
	},
}
