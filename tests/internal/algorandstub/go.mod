// Module algorandstub stands in for github.com/algorand/go-algorand so the
// msgp-generated *_gen_test.go files (which import that module's protocol and
// partitiontest packages) can compile and run inside this repo without taking
// a real dependency on go-algorand. It is wired in via a replace directive in
// ../../go.mod and is never published or imported by the msgp tool itself.
module github.com/algorand/go-algorand

go 1.23
