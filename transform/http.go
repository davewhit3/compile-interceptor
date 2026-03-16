package transform

import _ "embed"

//go:embed templates/http.go.tpl
var templateHttpSourceCode string

func init() {
	Register(&httpTransformer{
		transformer: &transformer{
			SourcePackage: "net/http",
			SourceFile:    "net/http/client.go",
			TemplateCode:  templateHttpSourceCode,
			TargetFunc:    "Do",
		},
	})
}

type httpTransformer struct {
	*transformer
}

func (t *httpTransformer) Support(importPath string) bool {
	return importPath == "net/http"
}
