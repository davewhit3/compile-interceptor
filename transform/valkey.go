package transform

import _ "embed"

//go:embed templates/valkey.go.tpl
var templateValkeySourceCode string

func init() {
	Register(&valkeyTransformer{
		transformer: &transformer{
			SourcePackage: "github.com/valkey-io/valkey-go",
			SourceFile:    "github.com/valkey-io/valkey-go@v1.0.73/client.go",
			TemplateCode:  templateValkeySourceCode,
			TargetFunc:    "Do",
			Imports:       []string{`"fmt"`},
		},
	})
}

type valkeyTransformer struct {
	*transformer
}
