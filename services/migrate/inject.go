package migrate

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/database/migration"
	"xorm.io/xorm"
	"xorm.io/xorm/migrate"
)

type Application struct {
	config config.Config
	logger zerolog.Logger
	db     *xorm.Engine
}

func (application Application) Start() error {
	m := migrate.New(application.db, &migrate.Options{
		TableName:    "migrations",
		IDColumnName: "id",
	}, migration.Migrations)
	if err := m.Migrate(); err != nil {
		return err
	}
	fmt.Println("migration complete")
	return nil
}

func newApplication(config config.Config, db *xorm.Engine, logger zerolog.Logger) Application {
	return Application{
		config: config,
		logger: logger,
		db:     db,
	}
}
