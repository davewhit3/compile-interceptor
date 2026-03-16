package compile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const pkgStoreDir = ".interceptor"

func ExtractImportCfgPath(args []string) string {
	for i, arg := range args {
		if arg == "-importcfg" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func ExtractOutputPath(args []string) string {
	for i, arg := range args {
		if arg == "-o" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// DeriveWorkDir returns the Go build work directory, trying the WORK env var
// first and falling back to deriving it from the -trimpath compiler arg.
func DeriveWorkDir(args []string) string {
	if w := os.Getenv("WORK"); w != "" {
		return w
	}
	trimpath := ExtractTrimpathFromArgs(args)
	if trimpath != "" {
		return filepath.Dir(trimpath)
	}
	return ""
}

// SavePkgArchivePath persists the compiled archive path for a package so that
// later interceptor invocations (separate processes) can retrieve it.
func SavePkgArchivePath(workDir, pkgImportPath, archivePath string) error {
	storeDir := filepath.Join(workDir, pkgStoreDir)
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return fmt.Errorf("creating store dir: %w", err)
	}

	filename := strings.ReplaceAll(pkgImportPath, "/", "_")
	return os.WriteFile(filepath.Join(storeDir, filename), []byte(archivePath), 0644)
}

// LoadPkgArchivePath reads the previously saved archive path for a package.
func LoadPkgArchivePath(workDir, pkgImportPath string) (string, error) {
	filename := strings.ReplaceAll(pkgImportPath, "/", "_")
	data, err := os.ReadFile(filepath.Join(workDir, pkgStoreDir, filename))
	if err != nil {
		return "", fmt.Errorf("reading package archive path for %s: %w", pkgImportPath, err)
	}
	return strings.TrimSpace(string(data)), nil
}

func AddPackage(importFilePath, pkgName, archivePath string) error {
	f, err := os.OpenFile(importFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening importcfg %s: %w", importFilePath, err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "packagefile %s=%s\n", pkgName, archivePath); err != nil {
		return fmt.Errorf("writing to importcfg %s: %w", importFilePath, err)
	}

	return nil
}
