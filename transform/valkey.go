package transform
import (_ "embed"; "regexp")
//go:embed templates/valkey.go.tpl
var templateValkeySourceCode string
var valkeyFilePattern = regexp.MustCompile(`github\.com/valkey-io/valkey-go@v[\d.]+/client\.go$`)
func init() {
	Register(&valkeyTransformer{transformer: &transformer{
		SourcePackage: "github.com/valkey-io/valkey-go",
		SourceFile: "client.go",
		SourceFilePattern: valkeyFilePattern,
		TemplateCode: templateValkeySourceCode,
		TargetFunc: "Do",
		Imports: []string{`"fmt"`},
	}})
}
type valkeyTransformer struct { *transformer }
