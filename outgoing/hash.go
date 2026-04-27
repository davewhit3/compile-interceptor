package outgoing

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
)

//go:embed *.go
var packageSource embed.FS

// SourceHash is a SHA-256 digest of all *.go files in the outgoing package,
// computed at build time via go:embed. Files are hashed in lexical order so
// the digest is deterministic. It is used by the interceptor's AlterToolVersion
// to include the injected dependency's content in the build cache key,
// preventing stale-cache fingerprint mismatches at link time.
var SourceHash = func() string {
	h := sha256.New()
	_ = fs.WalkDir(packageSource, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := packageSource.ReadFile(path)
		if err != nil {
			return err
		}
		h.Write(data)
		return nil
	})
	return fmt.Sprintf("%x", h.Sum(nil))
}()
