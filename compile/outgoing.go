package compile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// EnsurePkgArchive guarantees that a compiled archive of pkgPath exists in the
// current build's WORK dir and returns its path.
//
// Go's build scheduler compiles sibling dependencies in parallel, which means
// packages we want to transform may start compiling before an injected
// dependency is ready. This function side-builds pkgPath into
// workDir/.interceptor/<archiveName> using a plain `go build -o` (no
// -toolexec, so it does not recurse back into this interceptor). An flock on a
// sibling lockfile makes the build safe under parallel invocations.
func EnsurePkgArchive(workDir, pkgPath string) (string, error) {
	if workDir == "" {
		return "", fmt.Errorf("work dir is empty")
	}

	// Derive a stable archive filename from the last path segment.
	parts := strings.Split(pkgPath, "/")
	archiveName := parts[len(parts)-1] + ".a"

	storeDir := filepath.Join(workDir, pkgStoreDir)
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		return "", fmt.Errorf("creating store dir %s: %w", storeDir, err)
	}

	archivePath := filepath.Join(storeDir, archiveName)
	lockPath := archivePath + ".lock"

	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return "", fmt.Errorf("opening lock %s: %w", lockPath, err)
	}
	defer lock.Close()

	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return "", fmt.Errorf("locking %s: %w", lockPath, err)
	}
	defer func() { _ = syscall.Flock(int(lock.Fd()), syscall.LOCK_UN) }()

	// Another interceptor instance may have produced the archive while we
	// were waiting for the lock.
	if _, err := os.Stat(archivePath); err == nil {
		if err := SavePkgArchivePath(workDir, pkgPath, archivePath); err != nil {
			return "", fmt.Errorf("saving archive path: %w", err)
		}
		return archivePath, nil
	}

	tmpPath := archivePath + ".tmp"
	_ = os.Remove(tmpPath)

	cmd := exec.Command("go", "build", "-o", tmpPath, pkgPath)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("building %s: %w", pkgPath, err)
	}

	if err := os.Rename(tmpPath, archivePath); err != nil {
		return "", fmt.Errorf("renaming %s -> %s: %w", tmpPath, archivePath, err)
	}

	if err := SavePkgArchivePath(workDir, pkgPath, archivePath); err != nil {
		return "", fmt.Errorf("saving archive path: %w", err)
	}

	return archivePath, nil
}

// EnsureOutgoingArchive is a backward-compatible wrapper around EnsurePkgArchive.
func EnsureOutgoingArchive(workDir, outgoingPkg string) (string, error) {
	return EnsurePkgArchive(workDir, outgoingPkg)
}
