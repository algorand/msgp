// Package partitiontest is a minimal stand-in for
// github.com/algorand/go-algorand/test/partitiontest so generated tests that
// call partitiontest.PartitionTest(t) compile and run in this repo.
package partitiontest

import "testing"

// PartitionTest is a no-op stand-in for go-algorand's test-partitioning helper.
func PartitionTest(t *testing.T) {}
