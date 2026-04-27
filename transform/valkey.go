package transform

import (
	_ "embed"
	"regexp"
)

//go:embed templates/valkey.go.tpl
var templateValkeySourceCode string

// valkeyFilePattern matches valkey-go client.go with any version
// Examples it will match:
//   - github.com/valkey-io/valkey-go@v1.0.73/client.go
//   - github.com/valkey-io/valkey-go@v1.0.74/client.go
//   - github.com/valkey-io/valkey-go@v2.5.1/client.go
var valkeyFilePattern = regexp.MustCompile(`github\.com/valkey-io/valkey-go@v[\d.]+/client\.go$`)

func init() {
	Register(&valkeyTransformer{
		transformer: &transformer{
			SourcePackage:     "github.com/valkey-io/valkey-go",
			SourceFile:        "client.go",
			SourceFilePattern: valkeyFilePattern,
			TemplateCode:      templateValkeySourceCode,
			TargetFunc:        "Do",
			InjectedPkg:       "",
		},
	})
}

type valkeyTransformer struct {
	*transformer
}
