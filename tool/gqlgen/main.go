package main

import (
	"fmt"
	"os"

	"github.com/vektah/gqlparser/v2/ast"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/plugin/modelgen"
)

// Defining mutation function
func gormFieldHook(td *ast.Definition, fd *ast.FieldDefinition, f *modelgen.Field) (*modelgen.Field, error) {
	c := fd.Directives.ForName("gorm")
	if c != nil {
		format := c.Arguments.ForName("format")

		if format != nil {
			f.Tag += " gorm:" + format.Value.String()
		}

	}

	return f, nil
}

func main() {
	cfg, err := config.LoadConfigFromDefaultLocations()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config", err.Error())
		os.Exit(2)
	}

	// Attaching the mutation function onto modelgen plugin
	p := modelgen.Plugin{
		FieldHook: gormFieldHook,
	}

	err = api.Generate(cfg, api.ReplacePlugin(&p))

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(3)
	}
}
