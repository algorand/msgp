// This file exercises the maxtotalbytes directive, which gives a named slice
// type the same tight MaxSize bound that the maxtotalbytes= field tag gives a
// struct field.
package tests

//go:generate msgp

const (
	argCount   = 8
	argMaxSize = 100
	// argsTotal is much smaller than argCount*argMaxSize, so a MaxSize that
	// uses it is visibly tighter than the product of the allocbounds.
	argsTotal = 200
)

// BoundedArgs has only allocbounds, so its MaxSize is the product of them.
//
//msgp:allocbound BoundedArgs argCount,argMaxSize
type BoundedArgs [][]byte

// TotalBoundedArgs adds a maxtotalbytes directive whose bound is an
// expression containing spaces, to check that the whole expression survives.
//
//msgp:allocbound TotalBoundedArgs argCount,argMaxSize
//msgp:maxtotalbytes TotalBoundedArgs (argCount*msgp.BytesPrefixSize) + argsTotal
type TotalBoundedArgs [][]byte

// ArgsHolder holds TotalBoundedArgs, and should use its MaxSize.
type ArgsHolder struct {
	_struct struct{} `codec:",omitempty"`

	Args TotalBoundedArgs `codec:"a"`
}
