package outgoing

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
)

//go:embed store.go
var storeSource string

// SourceHash is a SHA-256 digest of the outgoing package source, computed at
// build time via go:embed. It is used by the interceptor's AlterToolVersion to
// include the injected dependency's content in the build cache key, preventing
// stale-cache fingerprint mismatches at link time.
var SourceHash = fmt.Sprintf("%x", sha256.Sum256([]byte(storeSource)))
