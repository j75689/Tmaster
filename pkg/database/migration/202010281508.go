package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/j75689/Tmaster/pkg/database/model"
	"gorm.io/gorm"
)

var v202010281508 = &gormigrate.Migration{
	ID: "202010281508",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Task{}); err != nil {
			return err
		}
		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		if err := tx.Migrator().DropTable(&model.Task{}); err != nil {
			return err
		}
		return nil
	},
}
