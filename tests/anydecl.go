// This file exercises top-level empty-interface type declarations: the
// generator must skip `type X any` exactly like `type X interface{}` (methods
// cannot be generated for an interface receiver, and the IDENT/Intf fallbacks
// emit code that does not compile), while ordinary types declared alongside
// them still generate.
package tests

//go:generate msgp

// HandleAny is an opaque reference in the style of go-algorand's
// agreement.MessageHandle and network.Peer; no methods may be generated.
type HandleAny any

// HandleAnyAlias is the alias form of the same declaration.
type HandleAnyAlias = any

// HandleIface is the pre-Go-1.18 spelling of HandleAny.
type HandleIface interface{}

// AnyDeclNeighbor must still get generated code despite the interface
// declarations surrounding it.
type AnyDeclNeighbor struct {
	_struct struct{} `codec:",omitempty"`

	Name  string `codec:"n"`
	Count uint64 `codec:"c"`
}
