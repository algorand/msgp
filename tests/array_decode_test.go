package tests

import (
	"testing"

	"github.com/algorand/msgp/msgp"
)

func wantArrayForMapError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected array-encoded map-struct to be rejected, got nil")
	}
	te, ok := msgp.Cause(err).(msgp.TypeError)
	if !ok {
		t.Fatalf("expected msgp.TypeError, got %T: %v", msgp.Cause(err), err)
	}
	if te.Method != msgp.MapType || te.Encoded != msgp.ArrayType {
		t.Fatalf("expected map-wanted/array-encoded TypeError, got %+v", te)
	}
}

func TestMapStructRejectsArray(t *testing.T) {
	// A single-field map-struct encoded as a one-element array.
	bts := msgp.AppendArrayHeader(nil, 1)
	bts = msgp.AppendString(bts, "hello")

	var out Inner
	_, err := out.UnmarshalMsg(bts)
	wantArrayForMapError(t, err)
	if out.X != "" {
		t.Fatalf("rejected decode must not populate fields, got X=%q", out.X)
	}
}

func TestMapStructRejectsMultiFieldArray(t *testing.T) {
	// An array whose elements line up positionally with Go field order.
	bts := msgp.AppendArrayHeader(nil, 2)
	bts = msgp.AppendString(bts, "somereq")
	bts = msgp.AppendString(bts, "someopt")

	var out MapRequiredOmitEmpty
	_, err := out.UnmarshalMsg(bts)
	wantArrayForMapError(t, err)
}

func TestMapStructRejectsEmptyArray(t *testing.T) {
	// A zero-length array. msgp once special-cased mfixarray(0) inside
	// ReadMapHeaderBytes as an empty map for go-codec parity; that special case
	// was folded into the now-removed fallback, so a zero-length array is a
	// plain map/array type mismatch today.
	bts := msgp.AppendArrayHeader(nil, 0)

	var out Inner
	_, err := out.UnmarshalMsg(bts)
	wantArrayForMapError(t, err)
}

func TestTupleStructStillDecodesArray(t *testing.T) {
	// Contrast: a struct that genuinely opts into array encoding via
	// //msgp:tuple still round-trips through its array form. Only the implicit
	// map-struct fallback was removed, not real tuple support.
	in := TupleRequired{A: "x", B: 7}
	var out TupleRequired
	if _, err := out.UnmarshalMsg(in.MarshalMsg(nil)); err != nil {
		t.Fatalf("expected tuple array decode to succeed, got %v", err)
	}
	if out != in {
		t.Fatalf("tuple round trip mismatch: got %+v, want %+v", out, in)
	}
}
