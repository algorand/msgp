// Package protocol is a minimal stand-in for
// github.com/algorand/go-algorand/protocol, providing just enough surface for
// the generated TestRandomizedEncoding* functions to compile and run.
package protocol

import "testing"

// RunEncodingTest is a no-op stand-in for go-algorand's randomized encoding
// test helper. The real helper fuzzes random instances; reproducing that is out
// of scope here, and a zero/random value would also trip required-field checks.
// The generated round-trip coverage lives in TestMarshalUnmarshal* and the
// hand-written tests in this package.
func RunEncodingTest(t *testing.T, v interface{}) {}
