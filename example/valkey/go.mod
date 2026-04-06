module github.com/davewhit3/compile-interceptor/example/valkey

go 1.25.0

replace github.com/davewhit3/compile-interceptor => ../../

require (
	github.com/davewhit3/compile-interceptor v0.0.0-00010101000000-000000000000
	github.com/valkey-io/valkey-go v1.0.73
)

require golang.org/x/sys v0.39.0 // indirect
