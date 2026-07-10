package parse

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/algorand/msgp/gen"
)

// TestGetTypeSpecsInterfaceDeclarations covers semantic classification of
// interface type declarations. Methods cannot be generated for an interface
// receiver, including when the interface is reached through an alias or
// another named type.
func TestGetTypeSpecsInterfaceDeclarations(t *testing.T) {
	src := `package p

type HandleAny any

type HandleAnyAlias = any

type HandleIface interface{}

type HandleNamed HandleIface

type HandleNamedAlias = HandleIface

type NonEmptyIface interface {
	Method()
}

type NonEmptyNamed NonEmptyIface

type Scalar uint64

type ScalarNamed Scalar

type ScalarAlias = Scalar

type Plain struct {
	A int
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	typeInfo := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
	if _, err = (&types.Config{}).Check("p", fset, []*ast.File{f}, typeInfo); err != nil {
		t.Fatal(err)
	}

	fs := &FileSet{
		Specs:      make(map[string]ast.Expr),
		Aliases:    make(map[string]ast.Expr),
		Interfaces: make(map[string]ast.Expr),
		Consts:     make(map[string]ast.Expr),
		Identities: make(map[string]gen.Elem),
	}
	fs.getTypeSpecs(f, typeInfo)

	interfaces := []string{
		"HandleAny",
		"HandleAnyAlias",
		"HandleIface",
		"HandleNamed",
		"HandleNamedAlias",
		"NonEmptyIface",
		"NonEmptyNamed",
	}
	for _, name := range interfaces {
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

	for _, name := range []string{"Scalar", "ScalarNamed", "Plain"} {
		if _, ok := fs.Specs[name]; !ok {
			t.Errorf("%s: expected in Specs", name)
		}
	}
	if _, ok := fs.Aliases["ScalarAlias"]; !ok {
		t.Errorf("ScalarAlias: expected in Aliases")
	}
}

func TestFileClassifiesAnyAsInterface(t *testing.T) {
	fs, err := File("./testdata/anydecl", false, "")
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"HandleAny", "HandleAnyAlias", "HandleIface", "HandleNamed"} {
		if _, ok := fs.Interfaces[name]; !ok {
			t.Errorf("%s: expected in Interfaces", name)
		}
		if _, ok := fs.Identities[name]; ok {
			t.Errorf("%s: must not be in Identities", name)
		}
	}
	if _, ok := fs.Identities["Neighbor"]; !ok {
		t.Error("Neighbor: expected in Identities")
	}
}
