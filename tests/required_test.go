package tests

import (
	"strings"
	"testing"

	"github.com/algorand/msgp/msgp"
)

// requiredCause reports the codec tag of the missing required field reported by
// err, or ("", false) if err is not a missing-required-field error. The
// generated decoder reports these as a plain wrapped error (no dedicated msgp
// type), so we match on the stable message prefix.
func requiredCause(err error) (string, bool) {
	const prefix = "missing required field: "
	if err == nil {
		return "", false
	}
	msg := msgp.Cause(err).Error()
	if !strings.HasPrefix(msg, prefix) {
		return "", false
	}
	return strings.TrimPrefix(msg, prefix), true
}

func TestRequiredMapDecodes(t *testing.T) {
	// Both required fields hold non-zero values: decode succeeds.
	in := MapRequired{ReqPlain: "x", ReqOmit: 7, Optional: "ignored"}
	bts := in.MarshalMsg(nil)

	var out MapRequired
	if _, err := out.UnmarshalMsg(bts); err != nil {
		t.Fatalf("expected clean decode, got %v", err)
	}
	if out != in {
		t.Fatalf("round trip mismatch: got %+v, want %+v", out, in)
	}
}

func TestRequiredMapMissingOmitted(t *testing.T) {
	// ReqOmit is required+omitempty: a zero value is dropped from the wire, so
	// decoding must report it as a missing required field. ReqPlain is set so
	// it is not the field that trips.
	in := MapRequired{ReqPlain: "x"}
	bts := in.MarshalMsg(nil)

	var out MapRequired
	_, err := out.UnmarshalMsg(bts)
	name, ok := requiredCause(err)
	if !ok || name != "reqomit" {
		t.Fatalf("expected ErrMissingRequiredField(reqomit), got %v", err)
	}
}

func TestRequiredMapPresentButZero(t *testing.T) {
	// ReqPlain is required without omitempty, and MapRequired's _struct has no
	// omitempty either, so a zero ReqPlain is written to the wire. It must
	// still be rejected on decode. ReqOmit is set so it does not trip first.
	in := MapRequired{ReqOmit: 1}
	bts := in.MarshalMsg(nil)

	var out MapRequired
	_, err := out.UnmarshalMsg(bts)
	name, ok := requiredCause(err)
	if !ok || name != "reqplain" {
		t.Fatalf("expected ErrMissingRequiredField(reqplain), got %v", err)
	}
}

func TestRequiredStructOmitEmpty(t *testing.T) {
	// _struct carries omitempty (the common convention), so every field is
	// omitempty. A non-zero required field round-trips cleanly.
	in := MapRequiredOmitEmpty{Req: "x", Optional: "y"}
	var out MapRequiredOmitEmpty
	if _, err := out.UnmarshalMsg(in.MarshalMsg(nil)); err != nil {
		t.Fatalf("expected clean decode, got %v", err)
	}
	if out != in {
		t.Fatalf("round trip mismatch: got %+v, want %+v", out, in)
	}

	// A zero required field is dropped by struct-level omitempty, so it is
	// absent on decode and must still be rejected.
	zero := MapRequiredOmitEmpty{Optional: "y"}
	_, err := (&MapRequiredOmitEmpty{}).UnmarshalMsg(zero.MarshalMsg(nil))
	if name, ok := requiredCause(err); !ok || name != "req" {
		t.Fatalf("expected ErrMissingRequiredField(req), got %v", err)
	}
}

func TestRequiredOptionalAbsentIsFine(t *testing.T) {
	// Only the required fields are set; the non-required Optional field is
	// absent. Decoding must succeed.
	in := MapRequired{ReqPlain: "x", ReqOmit: 1}
	bts := in.MarshalMsg(nil)

	var out MapRequired
	if _, err := out.UnmarshalMsg(bts); err != nil {
		t.Fatalf("expected clean decode with optional field absent, got %v", err)
	}
}

func TestRequiredComposite(t *testing.T) {
	// A composite required field with non-zero content: decode succeeds.
	in := CompositeRequired{Nested: Inner{X: "y"}}
	if _, err := (&CompositeRequired{}).UnmarshalMsg(in.MarshalMsg(nil)); err != nil {
		t.Fatalf("expected clean decode, got %v", err)
	}

	// A zero-valued nested struct (all fields zero) must be rejected,
	// exercising the type-derived zero check rather than a literal comparison.
	zeroNested := CompositeRequired{}
	_, err := (&CompositeRequired{}).UnmarshalMsg(zeroNested.MarshalMsg(nil))
	if name, ok := requiredCause(err); !ok || name != "nested" {
		t.Fatalf("expected ErrMissingRequiredField(nested), got %v", err)
	}
}

func TestRequiredTuple(t *testing.T) {
	// The required check also runs on the tuple (array) decode path.
	ok := TupleRequired{A: "present", B: 2}
	if _, err := (&TupleRequired{}).UnmarshalMsg(ok.MarshalMsg(nil)); err != nil {
		t.Fatalf("expected clean tuple decode, got %v", err)
	}

	zero := TupleRequired{B: 2} // A is required but zero
	_, err := (&TupleRequired{}).UnmarshalMsg(zero.MarshalMsg(nil))
	name, isReq := requiredCause(err)
	if !isReq || name != "a" {
		t.Fatalf("expected ErrMissingRequiredField(a), got %v", err)
	}
}
