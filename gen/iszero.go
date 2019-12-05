package gen

import (
	"io"
)

func isZeros(w io.Writer) *isZeroGen {
	return &isZeroGen{
		p: printer{w: w},
	}
}

type isZeroGen struct {
	passes
	p   printer
	ctx *Context
}

func (s *isZeroGen) Method() Method { return IsZero }

func (s *isZeroGen) Apply(dirs []string) error {
	return nil
}

func (s *isZeroGen) Execute(p Elem) error {
	if !s.p.ok() {
		return s.p.err
	}
	p = s.applyall(p)
	if p == nil {
		return nil
	}

	s.ctx = &Context{}
	s.ctx.PushString(p.TypeName())

	s.p.comment("MsgIsZero returns whether this is a zero value")

	if IsDangling(p) {
		p = p.Copy()
		baseType := p.(*BaseElem).IdentName
		ptrName := p.Varname()
		s.p.printf("\nfunc (%s %s) MsgIsZero() bool {", p.Varname(), methodReceiver(p))
		s.p.printf("\n  return ((*(%s))(%s)).MsgIsZero()", baseType, ptrName)
		s.p.printf("\n}")
		return s.p.err
	}

	ize := p.IfZeroExpr()
	if ize == "" {
		ize = "true"
	}
	s.p.printf("\nfunc (%s %s) MsgIsZero() bool { return %s }", p.Varname(), imutMethodReceiver(p), ize)
	return s.p.err
}
