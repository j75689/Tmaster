package wireset

import (
	"github.com/google/wire"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/database"
	"xorm.io/core"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

var DatabaseSet = wire.NewSet(
	InitializeDatabase,
)

func InitializeDatabase(config config.Config) (*xorm.Engine, error) {
	return database.NewDataBase(
		config.DB.Driver,
		config.DB.Host,
		config.DB.Port,
		config.DB.DBName,
		config.DB.InstanceName,
		config.DB.User,
		config.DB.Password,
		config.DB.SSLMode,
		config.DB.DialTimeout,
		database.SetConnMaxLifetime(config.DB.MaxLifetime),
		database.SetMaxIdleConns(config.DB.MaxIdleConn),
		database.SetMaxOpenConns(config.DB.MaxOpenConn),
		database.SetLogLevel(log.LogLevel(config.DB.LogLevel)),
		database.SetShowSQL(config.DB.ShowSQL),
		database.SetMapper(core.GonicMapper{}),
	)
}
