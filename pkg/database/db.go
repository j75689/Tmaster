package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// database driver for gorm
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
)

type DataSourceTypeName string

const (
	CloudMySql DataSourceTypeName = "cloud_mysql"
	Mysql      DataSourceTypeName = "mysql"
	Postgresql DataSourceTypeName = "postgres"
	Sqlite     DataSourceTypeName = "sqlite"
)

var _supportedDataSource = map[DataSourceTypeName]func(port uint, host, dbname, user, password, instanceName, connectTimeout, readTimeout, writeTimeout string, sslmode bool) gorm.Dialector{
	CloudMySql: func(port uint, host, dbname, user, password, instanceName, connectTimeout, readTimeout, writeTimeout string, sslmode bool) gorm.Dialector {
		return mysql.Open(fmt.Sprintf("%s:%s@unix(/cloudsql/%s)/%s?charset=utf8mb4&parseTime=true&loc=UTC&time_zone=UTC&timeout=%s&readTimeout=%s&writeTimeout=%s", user, password, instanceName, dbname, connectTimeout, readTimeout, writeTimeout))
	},
	Mysql: func(port uint, host, dbname, user, password, instanceName, connectTimeout, readTimeout, writeTimeout string, sslmode bool) gorm.Dialector {
		return mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=UTC&time_zone=UTC&timeout=%s&readTimeout=%s&writeTimeout=%s", user, password, host, port, dbname, connectTimeout, readTimeout, writeTimeout))
	},
	Postgresql: func(port uint, host, dbname, user, password, instanceName, connectTimeout, readTimeout, writeTimeout string, sslmode bool) gorm.Dialector {
		ssl := "disable"
		if sslmode {
			ssl = "allow"
		}
		return postgres.Open(fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s timezone=UTC", host, port, user, dbname, password, ssl))
	},
	Sqlite: func(port uint, host, dbname, user, password, instanceName, connectTimeout, readTimeout, writeTimeout string, sslmode bool) gorm.Dialector {
		return sqlite.Open(fmt.Sprintf("%s.db", dbname))
	},
}

func NewDataBase(
	driver, host string,
	port uint, dbname, instanceName string,
	user, password string, sslmode bool,
	connectTimeout, readTimeout, writeTimeout string,
	dialTimeout time.Duration,
	options ...Option,
) (*gorm.DB, error) {
	supported, ok := _supportedDataSource[DataSourceTypeName(driver)]
	if !ok {
		return nil, fmt.Errorf("unsupported sql driver [%s]", driver)
	}
	sqlDriver := supported(
		port, host, dbname,
		user, password, instanceName,
		connectTimeout, readTimeout, writeTimeout, sslmode)

	engine, err := gorm.Open(sqlDriver, &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	sql, err := engine.DB()
	if err != nil {
		return nil, err
	}
	err = sql.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	for _, opt := range options {
		err = opt.Apply(engine)
		if err != nil {
			return nil, err
		}
	}
	return engine, nil
}

// An Option configures a gorm.DB
type Option interface {
	Apply(*gorm.DB) error
}

// OptionFunc is a function that configures a gorm.DB
type OptionFunc func(*gorm.DB) error

// Apply is a function that set value to gorm.DB
func (f OptionFunc) Apply(engine *gorm.DB) error {
	return f(engine)
}

func SetConnMaxIdleTime(maxIdleTime time.Duration) Option {
	return OptionFunc(func(engine *gorm.DB) error {
		sql, err := engine.DB()
		if err != nil {
			return err
		}
		sql.SetConnMaxIdleTime(maxIdleTime)
		return nil
	})
}

func SetConnMaxLifetime(maxlifetime time.Duration) Option {
	return OptionFunc(func(engine *gorm.DB) error {
		sql, err := engine.DB()
		if err != nil {
			return err
		}
		sql.SetConnMaxLifetime(maxlifetime)
		return nil
	})
}

func SetMaxIdleConns(maxIdleConns int) Option {
	return OptionFunc(func(engine *gorm.DB) error {
		sql, err := engine.DB()
		if err != nil {
			return err
		}
		sql.SetMaxIdleConns(maxIdleConns)
		return nil
	})
}

func SetMaxOpenConns(maxOpenConns int) Option {
	return OptionFunc(func(engine *gorm.DB) error {
		sql, err := engine.DB()
		if err != nil {
			return err
		}
		sql.SetMaxOpenConns(maxOpenConns)
		return nil
	})
}

func SetLogLevel(logLevel logger.LogLevel) Option {
	return OptionFunc(func(engine *gorm.DB) error {
		engine.Logger = engine.Logger.LogMode(logLevel)
		return nil
	})
}

func SetLogger(logger logger.Interface) Option {
	return OptionFunc(func(engine *gorm.DB) error {
		engine.Logger = logger
		return nil
	})
}
