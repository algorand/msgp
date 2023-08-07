package parse

import (
	"errors"
	"fmt"
	"go/ast"
	"strings"

	"github.com/algorand/msgp/gen"
)

const linePrefix = "//msgp:"

// func(args, fileset)
type directive func([]string, *FileSet) error

// func(passName, args, printer)
type passDirective func(gen.Method, []string, *gen.Printer) error

// map of all recognized directives
//
// to add a directive, define a func([]string, *FileSet) error
// and then add it to this list.
var directives = map[string]directive{
	"shim":       applyShim,
	"ignore":     ignore,
	"tuple":      astuple,
	"sort":       sortintf,
	"allocbound": allocbound,
	// _postunmarshalcheck is used to add callbacks to the end of un-marshalling that are tied to a specific Element.
	_postunmarshalcheck: postunmarshalcheck,
}

// map of base types with predefined LessFunctions used in the sort directive
var lessFns = map[string]string{
	"int":     "msgp.IntLess",
	"int8":    "msgp.Int8Less",
	"int16":   "msgp.Int16Less",
	"int32":   "msgp.Int32Less",
	"int64":   "msgp.Int64Less",
	"uint":    "msgp.UintLess",
	"uint8":   "msgp.Uint8Less",
	"uint16":  "msgp.Uint16Less",
	"uint32":  "msgp.Uint32Less",
	"uint64":  "msgp.Uint64Less",
	"float32": "msgp.Float32Less",
	"float64": "msgp.Float64Less",
	"string":  "msgp.StringLess",
	"[]byte":  "msgp.BytesLess",
}

const _postunmarshalcheck = "postunmarshalcheck"

var errNotEnoughArguments = errors.New("postunmarshalcheck did not receive enough arguments. expected at least 3")

// the functions should have no params, and output zero.
//
//msgp:postunmarshalcheck {Type} {funcName} {funcName} ...
func postunmarshalcheck(text []string, f *FileSet) error {
	if len(text) < 3 {
		return errNotEnoughArguments
	}
	// not error but doesn't do anything
	if text[0] != _postunmarshalcheck {
		return nil
	}
	text = text[1:]

	elemType := text[0]
	elem, ok := f.Identities[elemType]
	if !ok {
		return errors.New(fmt.Sprintf("postunmarshalcheck error: type %v does not exist", elemType))
	}
	for _, fName := range text[1:] {
		elem.AddCallback(gen.Callback{
			Fname:        fName,
			CallbackType: gen.UnmarshalCallBack,
		})
	}
	return nil
}

var passDirectives = map[string]passDirective{
	"ignore": passignore,
}

func passignore(m gen.Method, text []string, p *gen.Printer) error {
	pushstate(m.String())
	for _, a := range text {
		p.ApplyDirective(m, gen.IgnoreTypename(a))
		infof("ignoring %s\n", a)
	}
	popstate()
	return nil
}

// find all comment lines that begin with //msgp:
func yieldComments(c []*ast.CommentGroup) []string {
	var out []string
	for _, cg := range c {
		for _, line := range cg.List {
			if strings.HasPrefix(line.Text, linePrefix) {
				out = append(out, strings.TrimPrefix(line.Text, linePrefix))
			}
		}
	}
	return out
}

//msgp:shim {Type} as:{Newtype} using:{toFunc/fromFunc} mode:{Mode}
func applyShim(text []string, f *FileSet) error {
	if len(text) < 4 || len(text) > 5 {
		return fmt.Errorf("shim directive should have 3 or 4 arguments; found %d", len(text)-1)
	}

	name := text[1]
	be := gen.Ident("", strings.TrimPrefix(strings.TrimSpace(text[2]), "as:")) // parse as::{base}
	if name[0] == '*' {
		name = name[1:]
		be.Needsref(true)
	}
	be.Alias(name)

	usestr := strings.TrimPrefix(strings.TrimSpace(text[3]), "using:") // parse using::{method/method}

	methods := strings.Split(usestr, "/")
	if len(methods) != 2 {
		return fmt.Errorf("expected 2 using::{} methods; found %d (%q)", len(methods), text[3])
	}

	be.ShimToBase = methods[0]
	be.ShimFromBase = methods[1]

	if len(text) == 5 {
		modestr := strings.TrimPrefix(strings.TrimSpace(text[4]), "mode:") // parse mode::{mode}
		switch modestr {
		case "cast":
			be.ShimMode = gen.Cast
		case "convert":
			be.ShimMode = gen.Convert
		default:
			return fmt.Errorf("invalid shim mode; found %s, expected 'cast' or 'convert", modestr)
		}
	}

	infof("%s -> %s\n", name, be.Value.String())
	f.findShim(name, be)

	return nil
}

//msgp:ignore {TypeA} {TypeB}...
func ignore(text []string, f *FileSet) error {
	if len(text) < 2 {
		return nil
	}
	for _, item := range text[1:] {
		name := strings.TrimSpace(item)
		if _, ok := f.Identities[name]; ok {
			delete(f.Identities, name)
			infof("ignoring %s\n", name)
		}
	}
	return nil
}

//msgp:tuple {TypeA} {TypeB}...
func astuple(text []string, f *FileSet) error {
	if len(text) < 2 {
		return nil
	}
	for _, item := range text[1:] {
		name := strings.TrimSpace(item)
		if el, ok := f.Identities[name]; ok {
			if st, ok := el.(*gen.Struct); ok {
				st.AsTuple = true
				infoln(name)
			} else {
				warnf("%s: only structs can be tuples\n", name)
			}
		}
	}
	return nil
}

//msgp:sort {Type} {SortInterface} {LessFunction}
func sortintf(text []string, f *FileSet) error {
	if len(text) != 4 && len(text) != 3 {
		return nil
	}
	sortType := strings.TrimSpace(text[1])
	sortIntf := strings.TrimSpace(text[2])
	gen.SetSortInterface(sortType, sortIntf)
	infof("sorting %s using %s\n", sortType, sortIntf)
	var lessFn string
	if len(text) == 4 {
		lessFn = strings.TrimSpace(text[3])
	} else if fn, ok := lessFns[sortType]; ok {
		lessFn = fn
	} else {
		panic(fmt.Sprintf("no default less function for %s and no function is provided", sortType))
	}
	gen.SetLessFunction(sortType, lessFn)
	infof("less fn %s using %s\n", sortType, lessFn)
	return nil
}

//msgp:allocbound {Type} {Bound}
func allocbound(text []string, f *FileSet) error {
	if len(text) != 3 {
		return nil
	}
	allocBoundType := strings.TrimSpace(text[1])
	allocBound := strings.TrimSpace(text[2])
	t, ok := f.Identities[allocBoundType]
	if !ok {
		warnf("allocbound: cannot find type %s\n", allocBoundType)
	} else {
		t.SetAllocBound(allocBound)
		infof("allocbound(%s): setting to %s\n", allocBoundType, allocBound)
	}
	return nil
}
