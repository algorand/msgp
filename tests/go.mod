// This is a nested module so the msgp-generated *_gen_test.go files, which
// import go-algorand's protocol and partitiontest packages, can compile and run
// against local stubs. Keeping it separate from the root module means the root
// go.mod stays free of replace directives (which would break
// `go run github.com/algorand/msgp@version`) and free of a go-algorand
// dependency. Run these tests with: cd tests && go test ./...
module github.com/algorand/msgp/tests

go 1.25.0

require (
	github.com/algorand/go-algorand v0.0.0-00010101000000-000000000000
	github.com/algorand/msgp v0.0.0
)

replace github.com/algorand/msgp => ../

replace github.com/algorand/go-algorand => ./internal/algorandstub
