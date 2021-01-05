package wireset

import (
	"github.com/google/wire"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/database"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DatabaseSet = wire.NewSet(
	InitializeDatabase,
)

func InitializeDatabase(config config.Config) (*gorm.DB, error) {
	return database.NewDataBase(
		config.DB.Driver,
		config.DB.Host,
		config.DB.Port,
		config.DB.DBName,
		config.DB.InstanceName,
		config.DB.User,
		config.DB.Password,
		config.DB.SSLMode,
		config.DB.ConnectTimeout,
		config.DB.ReadTimeout,
		config.DB.WriteTimeout,
		config.DB.DialTimeout,
		database.SetConnMaxLifetime(config.DB.MaxLifetime),
		database.SetMaxIdleConns(config.DB.MaxIdleConn),
		database.SetMaxOpenConns(config.DB.MaxOpenConn),
		database.SetConnMaxIdleTime(config.DB.MaxIdleTime),
		database.SetLogLevel(logger.LogLevel(config.DB.LogLevel)),
	)
}
