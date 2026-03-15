# compile-interceptor

Go compile-time code instrumentation tool that uses `-toolexec` to intercept the Go compiler and apply AST transformations to standard library packages during build.

Currently ships with an `net/http` transformer that injects request duration logging into `(*Client).Do`.

## How it works

1. Go's `-toolexec` flag lets you wrap every tool invocation (compile, link, asm, etc.) with a custom binary.
2. When the Go toolchain calls `compile` for a target package (e.g. `net/http`), the interceptor:
   - Parses the original source file using [dave/dst](https://github.com/dave/dst) (Decorated Syntax Tree).
   - Loads a template with the replacement function body.
   - Swaps the target function's body with the template version.
   - Writes a modified `.go` file and passes it to the real compiler instead of the original.
3. For all other tools (link, asm, etc.) the interceptor proxies the call unchanged.
4. Cache busting is handled automatically — the interceptor modifies `-V=full` output so `go build` doesn't serve stale cached artifacts.


## Requirements

- Go 1.25+
- macOS arm64 (paths in Makefile assume Homebrew Go; adjust `COMPILER` for your platform)

## Build

```bash
make build
```

This produces `dist/interceptor`.

## Usage

Build any Go program with the interceptor as `-toolexec`:

```bash
go build -toolexec '/path/to/dist/interceptor' -o myapp ./cmd/myapp
```

Or use the provided Makefile targets with the included example app:

```bash
# build interceptor, then compile example/main.go with instrumentation
make build && make test
```

The `test` target runs:

```bash
go build -work -toolexec './dist/interceptor' -o dist/example example/main.go
```

After building, run the instrumented binary:

```bash
./dist/example
```

You should see the injected duration output for every `http.Client.Do` call:

```
Hello, World!
duration 123.456ms
Response: 200
```

## Local debugging

The `local` target lets you invoke the interceptor directly against a single `compile` command, useful for debugging transformations without a full build:

```bash
make build && make local
```

This simulates the compiler invocation for `net/http` using artifacts from `testdata/`.

## Running tests

```bash
make tests
```

## Adding a new transformer

1. Create a new file in `transform/` (e.g. `transform/dns.go`).
2. Define a struct embedding `*transformer` and implement the `Support(importPath string) bool` method.
3. Create a template file in `transform/templates/` with the replacement function body.
4. Register the transformer in an `init()` function using `Register(...)`.

Example:

```go
package transform

import _ "embed"

//go:embed templates/dns.go.tpl
var templateDnsSourceCode string

func init() {
    Register(&dnsTransformer{
        transformer: &transformer{
            SourceFile:   "net/dns/lookup.go",
            TemplateCode: templateDnsSourceCode,
            TargetFunc:   "LookupHost",
        },
    })
}

type dnsTransformer struct {
    *transformer
}

func (t *dnsTransformer) Support(importPath string) bool {
    return importPath == "net/dns"
}
```

## Environment variables

| Variable | Description |
|---|---|
| `TOOLEXEC_IMPORTPATH` | Set automatically by `go build -toolexec`; contains the import path of the package being compiled. |
| `WORK` | Build work directory; set automatically by `go build -work` or manually for `make local`. |

## License

See [LICENSE](LICENSE) for details.