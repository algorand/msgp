package tests

import (
	"testing"

	"github.com/algorand/msgp/msgp"
)

func TestMaxTotalBytesDirective(t *testing.T) {
	// Without the directive, MaxSize assumes every slot is full and maximal.
	product := msgp.ArrayHeaderSize + argCount*(msgp.BytesPrefixSize+argMaxSize)
	if got := BoundedArgsMaxSize(); got != product {
		t.Fatalf("BoundedArgsMaxSize() = %d, want %d", got, product)
	}

	// With it, MaxSize is exactly the directive's expression plus the header.
	total := msgp.ArrayHeaderSize + argCount*msgp.BytesPrefixSize + argsTotal
	if got := TotalBoundedArgsMaxSize(); got != total {
		t.Fatalf("TotalBoundedArgsMaxSize() = %d, want %d", got, total)
	}
	if total >= product {
		t.Fatalf("test constants should make the total bound (%d) tighter than the product (%d)", total, product)
	}

	// A struct holding the type picks up the same tight bound, plus a
	// one-byte map header and the two-byte key "a".
	holder := 1 + 2 + total
	if got := ArgsHolderMaxSize(); got != holder {
		t.Fatalf("ArgsHolderMaxSize() = %d, want %d", got, holder)
	}
}

func TestMaxTotalBytesDirectiveIsBound(t *testing.T) {
	// The most bytes a value obeying argCount and argsTotal can encode to:
	// every slot used, with the total spread so each element needs a
	// multi-byte bin header.
	var args TotalBoundedArgs
	for i := 0; i < argCount; i++ {
		args = append(args, make([]byte, argsTotal/argCount))
	}
	if n := len(args.MarshalMsg(nil)); n > TotalBoundedArgsMaxSize() {
		t.Fatalf("encoded %d bytes, more than TotalBoundedArgsMaxSize() = %d", n, TotalBoundedArgsMaxSize())
	}
	h := ArgsHolder{Args: args}
	if n := len(h.MarshalMsg(nil)); n > ArgsHolderMaxSize() {
		t.Fatalf("encoded %d bytes, more than ArgsHolderMaxSize() = %d", n, ArgsHolderMaxSize())
	}

	// The directive only affects MaxSize; decoding still enforces the
	// per-element allocbound.
	over := TotalBoundedArgs{make([]byte, argMaxSize+1)}
	var out TotalBoundedArgs
	if _, err := out.UnmarshalMsg(over.MarshalMsg(nil)); err == nil {
		t.Fatalf("decoded an element longer than argMaxSize")
	}
}
