package migrate

import (
	"fmt"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/database/migration"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Application struct {
	config config.Config
	logger zerolog.Logger
	db     *gorm.DB
}

func (application Application) Start() error {
	m := gormigrate.New(application.db, gormigrate.DefaultOptions, migration.Migrations)
	if err := m.Migrate(); err != nil {
		return err
	}
	fmt.Println("migration complete")
	return nil
}

func newApplication(config config.Config, db *gorm.DB, logger zerolog.Logger) Application {
	return Application{
		config: config,
		logger: logger,
		db:     db,
	}
}
