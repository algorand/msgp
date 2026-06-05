// Package tests exercises the `required` codec struct-tag option, which makes
// the generated decoder reject any field that holds its zero value after
// decoding (whether the field was absent from the wire or encoded as zero).
package tests

//go:generate msgp -tests=false

//msgp:tuple TupleRequired

// MapRequired covers the required option on a map-encoded struct. Its _struct
// tag carries no omitempty, so a zero required field is still written to the
// wire (the "present but zero" rejection path); a field that individually opts
// into omitempty is instead dropped when zero (the "absent from the wire"
// path).
type MapRequired struct {
	_struct struct{} `codec:""`

	// Required, not omitempty: a zero value is written to the wire and must
	// still be rejected on decode.
	ReqPlain string `codec:"reqplain,required"`

	// Required and omitempty: a zero value is dropped by the encoder, leaving
	// the field absent on decode, and must still be rejected.
	ReqOmit int64 `codec:"reqomit,omitempty,required"`

	// Not required: a zero value decodes without error.
	Optional string `codec:"opt,omitempty"`
}

// MapRequiredOmitEmpty has the same shape as MapRequired but its _struct tag
// carries omitempty, so struct-level omitempty applies to every field. A zero
// required field is dropped from the wire entirely (rather than written as
// zero) and must still be rejected on decode.
type MapRequiredOmitEmpty struct {
	_struct struct{} `codec:",omitempty"`

	// Required with struct-level (not field-level) omitempty: a zero value is
	// still dropped by the encoder, leaving the field absent on decode.
	Req string `codec:"req,required"`

	// Not required.
	Optional string `codec:"opt"`
}

// Inner is a named struct used to exercise required on composite fields.
type Inner struct {
	_struct struct{} `codec:",omitempty"`
	X       string   `codec:"x,omitempty"`
}

// CompositeRequired covers required on a non-scalar field, whose zero check is
// derived from the type rather than a literal: a value struct is zero when all
// of its fields are zero.
type CompositeRequired struct {
	_struct struct{} `codec:",omitempty"`

	Nested Inner `codec:"nested,required"` // zero when Inner is zero (X == "")
}

// TupleRequired is an array-encoded struct (see the //msgp:tuple directive)
// confirming the required check also runs on the tuple decode path.
type TupleRequired struct {
	A string `codec:"a,required"`
	B int64  `codec:"b"`
}
