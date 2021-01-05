package migration

import "github.com/go-gormigrate/gormigrate/v2"

// Version is a migrate version of database
type Version struct {
	ID   int64
	Name string
}

// Migrations is a collection of storage migration patterns
var Migrations = []*gormigrate.Migration{
	v202004271600,
	v202010281508,
}
