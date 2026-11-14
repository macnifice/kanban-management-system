package internal_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// Regla de dependencias de DDD: las capas internas no conocen a las externas.
//
//	domain  ←  application  ←  infrastructure / interfaces
var forbidden = map[string][]string{
	"domain": {
		"kanban/internal/application", "kanban/internal/infrastructure", "kanban/internal/interfaces",
		"database/sql", "net/http", "github.com/",
	},
	"application": {
		"kanban/internal/infrastructure", "kanban/internal/interfaces", "database/sql", "net/http", "github.com/",
	},
	"infrastructure": {"kanban/internal/interfaces"},
	"interfaces":     {"kanban/internal/infrastructure", "database/sql"},
}

func TestLayerDependencies(t *testing.T) {
	for layer, banned := range forbidden {
		err := filepath.WalkDir(layer, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			f, perr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if perr != nil {
				return perr
			}
			for _, imp := range f.Imports {
				p := strings.Trim(imp.Path.Value, `"`)
				for _, b := range banned {
					if p == b || strings.HasPrefix(p, b) {
						t.Errorf("%s importa %q: la capa %q no puede depender de eso", path, p, layer)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
