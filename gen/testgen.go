package gen

import (
	"go/ast"
	"io"
	"text/template"
)

var (
	marshalTestTempl = template.New("MarshalTest")
)

// TODO(philhofer):
// for simplicity's sake, right now
// we can only generate tests for types
// that can be initialized with the
// "Type{}" syntax.
// we should support all the types.

func mtest(w io.Writer) *mtestGen {
	return &mtestGen{w: w}
}

type mtestGen struct {
	passes
	w io.Writer
}

// mtestData is the data handed to the marshal-test template.
type mtestData struct {
	TypeName string
	// HasRequired is true when decoding the zero value of this type would trip
	// a generated `required` field check and return an error. The zero-value
	// round-trip assertions are relaxed in that case, since a zero value is by
	// definition not a valid encoding of a type with required fields.
	HasRequired bool
}

func (m *mtestGen) Execute(p Elem) ([]string, error) {
	p = m.applyall(p)
	if p != nil && !IsDangling(p) {
		switch p.(type) {
		case *Struct, *Array, *Slice, *Map:
			return nil, marshalTestTempl.Execute(m.w, mtestData{
				TypeName:    p.TypeName(),
				HasRequired: zeroValueHasRequired(p),
			})
		}
	}
	return nil, nil
}

func (m *mtestGen) Method() Method { return marshaltest }

// zeroValueHasRequired reports whether decoding the zero value of e would fail
// a generated `required` check. Only exported fields get checks. A zero slice
// or map is empty, so its elements (and their checks) are never decoded; a zero
// array decodes its zero elements, so its element type still matters.
func zeroValueHasRequired(e Elem) bool {
	switch e := e.(type) {
	case *Struct:
		for i := range e.Fields {
			if ast.IsExported(e.Fields[i].FieldName) && e.Fields[i].HasTagPart("required") {
				return true
			}
		}
	case *Array:
		return zeroValueHasRequired(e.Els)
	}
	return false
}

func init() {
	template.Must(marshalTestTempl.Parse(`func TestMarshalUnmarshal{{.TypeName}}(t *testing.T) {
	partitiontest.PartitionTest(t)
	v := {{.TypeName}}{}
	bts := v.MarshalMsg(nil)
	left, err := v.UnmarshalMsg(bts)
{{if .HasRequired}}	// The zero value omits its required field(s), so UnmarshalMsg is expected to
	// reject it; only MarshalMsg and Skip are exercised against a zero value.
	if err == nil {
		t.Errorf("expected a missing-required-field error decoding a zero {{.TypeName}}")
	}
{{else}}	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after UnmarshalMsg(): %q", len(left), left)
	}
{{end}}
	left, err = msgp.Skip(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after Skip(): %q", len(left), left)
	}
}

func TestRandomizedEncoding{{.TypeName}}(t *testing.T) {
	protocol.RunEncodingTest(t, &{{.TypeName}}{})
}

func BenchmarkMarshalMsg{{.TypeName}}(b *testing.B) {
	v := {{.TypeName}}{}
	b.ReportAllocs()
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		v.MarshalMsg(nil)
	}
}

func BenchmarkAppendMsg{{.TypeName}}(b *testing.B) {
	v := {{.TypeName}}{}
	bts := make([]byte, 0, v.Msgsize())
	bts = v.MarshalMsg(bts[0:0])
	b.SetBytes(int64(len(bts)))
	b.ReportAllocs()
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		bts = v.MarshalMsg(bts[0:0])
	}
}

func BenchmarkUnmarshal{{.TypeName}}(b *testing.B) {
	v := {{.TypeName}}{}
	bts := v.MarshalMsg(nil)
	b.ReportAllocs()
	b.SetBytes(int64(len(bts)))
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
{{if .HasRequired}}		// The zero value fails the required-field check; ignore the error so the
		// benchmark still measures the decoding work.
		_, _ = v.UnmarshalMsg(bts)
{{else}}		_, err := v.UnmarshalMsg(bts)
		if err != nil {
			b.Fatal(err)
		}
{{end}}	}
}

`))

}
