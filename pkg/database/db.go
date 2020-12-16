package database

import (
	"context"
	"fmt"
	"time"

	"xorm.io/xorm"
	"xorm.io/xorm/log"
	"xorm.io/xorm/names"

	// sql driver
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type DataSourceTypeName string

const (
	CloudMySql DataSourceTypeName = "cloud_mysql"
	Mysql      DataSourceTypeName = "mysql"
	Postgresql DataSourceTypeName = "postgres"
	Sqlite     DataSourceTypeName = "sqlite"
)

var _supportedDataSource = map[DataSourceTypeName]func(driver DataSourceTypeName, host string, port uint, dbname string, instanceName string, user string, password string, sslMode bool) (*xorm.Engine, error){
	CloudMySql: func(driver DataSourceTypeName, host string, port uint, dbname string, instanceName string, user string, password string, sslMode bool) (*xorm.Engine, error) {
		return xorm.NewEngine(
			"mysql",
			fmt.Sprintf("%s:%s@unix(/cloudsql/%s)/%s?charset=utf8mb4&parseTime=true&loc=UTC&time_zone=UTC",
				user, password, instanceName, dbname))
	},
	Mysql: func(driver DataSourceTypeName, host string, port uint, dbname string, instanceName string, user string, password string, sslMode bool) (*xorm.Engine, error) {
		return xorm.NewEngine(
			"mysql",
			fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=UTC&time_zone=UTC",
				user, password, host, port, dbname))
	},
	Postgresql: func(driver DataSourceTypeName, host string, port uint, dbname string, instanceName string, user string, password string, sslMode bool) (*xorm.Engine, error) {
		ssl := "disable"
		if sslMode {
			ssl = "allow"
		}
		return xorm.NewEngine(
			"postgres",
			fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s timezone=UTC",
				host, port, user, dbname, password, ssl))
	},
	Sqlite: func(driver DataSourceTypeName, host string, port uint, dbname string, instanceName string, user string, password string, sslMode bool) (*xorm.Engine, error) {
		return xorm.NewEngine("sqlite", fmt.Sprintf("%s.db", dbname))
	},
}

func NewDataBase(
	driver string,
	host string,
	port uint,
	dbname string,
	instanceName string,
	user string,
	password string,
	sslMode bool,
	dialTimeout time.Duration,
	options ...Option,
) (*xorm.Engine, error) {
	supported, ok := _supportedDataSource[DataSourceTypeName(driver)]
	if !ok {
		return nil, fmt.Errorf("unsupported driver [%s]", driver)
	}

	engine, err := supported(DataSourceTypeName(driver), host, port, dbname, instanceName, user, password, sslMode)
	if err != nil {
		return nil, err
	}
	// set value
	for _, option := range options {
		option.Apply(engine)
	}
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()
	if err := engine.Context(ctx).Ping(); err != nil {
		return nil, err
	}
	return engine, nil
}

// An Option configures a xorm.Engine
type Option interface {
	Apply(*xorm.Engine)
}

// OptionFunc is a function that configures a xorm.Engine
type OptionFunc func(*xorm.Engine)

// Apply is a function that set value to xorm.Engine
func (f OptionFunc) Apply(engine *xorm.Engine) {
	f(engine)
}

func SetConnMaxLifetime(maxlifetime time.Duration) Option {
	return OptionFunc(func(engine *xorm.Engine) {
		engine.SetConnMaxLifetime(maxlifetime)
	})
}

func SetMaxIdleConns(maxIdleConns int) Option {
	return OptionFunc(func(engine *xorm.Engine) {
		engine.SetMaxIdleConns(maxIdleConns)
	})
}

func SetMaxOpenConns(maxOpenConns int) Option {
	return OptionFunc(func(engine *xorm.Engine) {
		engine.SetMaxOpenConns(maxOpenConns)
	})
}

func SetLogLevel(logLevel log.LogLevel) Option {
	return OptionFunc(func(engine *xorm.Engine) {
		engine.SetLogLevel(logLevel)
	})
}

func SetLogger(logger interface{}) Option {
	return OptionFunc(func(engine *xorm.Engine) {
		engine.SetLogger(logger)
	})
}

func SetShowSQL(showSQL ...bool) Option {
	return OptionFunc(func(engine *xorm.Engine) {
		engine.ShowSQL(showSQL...)
	})
}

func SetMapper(mapper names.Mapper) Option {
	return OptionFunc(func(engine *xorm.Engine) {
		engine.SetMapper(mapper)
	})
}
