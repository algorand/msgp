package parse

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/algorand/msgp/gen"
)

// TestGetTypeSpecsAnyDeclaration covers classification of empty-interface
// type declarations. `type X interface{}` is routed to fs.Interfaces so no
// methods are generated for it; `type X any` must be routed the same way,
// since methods cannot be generated for an interface receiver (and the
// fallback used to emit non-compiling code for such declarations).
func TestGetTypeSpecsAnyDeclaration(t *testing.T) {
	src := `package p

type HandleAny any

type HandleAnyAlias = any

type HandleIface interface{}

type Plain struct {
	A int
}
`
	f, err := parser.ParseFile(token.NewFileSet(), "p.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	fs := &FileSet{
		Specs:      make(map[string]ast.Expr),
		Aliases:    make(map[string]ast.Expr),
		Interfaces: make(map[string]ast.Expr),
		Consts:     make(map[string]ast.Expr),
		Identities: make(map[string]gen.Elem),
	}
	fs.getTypeSpecs(f)

	for _, name := range []string{"HandleAny", "HandleAnyAlias", "HandleIface"} {
		if _, ok := fs.Interfaces[name]; !ok {
			t.Errorf("%s: expected in Interfaces", name)
		}
		if _, ok := fs.Specs[name]; ok {
			t.Errorf("%s: must not be in Specs (would generate methods on an interface)", name)
		}
		if _, ok := fs.Aliases[name]; ok {
			t.Errorf("%s: must not be in Aliases", name)
		}
	}

	if _, ok := fs.Specs["Plain"]; !ok {
		t.Errorf("Plain: expected in Specs")
	}
}
