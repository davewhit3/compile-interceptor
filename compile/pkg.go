package compile

import (
	"fmt"
	"slices"
)

// ExtractFilesFromPack locates the -pack flag in args, and returns the Go source files listed after it.
// It also returns the index offset at which the Go files begin in the original args slice.
func ExtractFilesFromPack(args []string) ([]string, int, error) {
	packIndex := slices.Index(args, "-pack")

	if packIndex == -1 {
		return nil, 0, fmt.Errorf("-pack flag is not found")
	}

	filesCount := len(args) - packIndex
	files := make([]string, filesCount)
	goFiles := args[packIndex+1:]
	copy(files, goFiles)

	goFilesIndex := packIndex + ExecCmdArgsOffset + 1

	return goFiles, goFilesIndex, nil
}
