package main

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/davewhit3/compile-interceptor/compile"
	"github.com/davewhit3/compile-interceptor/transform"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	log := slog.Default()

	importPath := os.Getenv("TOOLEXEC_IMPORTPATH")
	workDir := os.Getenv("WORK")

	args := os.Args[compile.ExecCmdArgsOffset:]
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: compile-interceptor <compiler> [args...]\n")
		os.Exit(1)
	}

	tool := args[0]
	toolArgs := args[1:]
	toolName := filepath.Base(tool)
	trimpath := compile.ExtractTrimpathFromArgs(toolArgs)

	log.Debug(
		"context",
		"trimpath", trimpath,
		"tool", toolName,
		"workDir", workDir,
		"importPath", importPath,
	)

	if compile.IsCacheBuild(os.Args) {
		if err := compile.AlterToolVersion(tool, toolArgs); err != nil {
			log.Error("alter tool version failed", "err", err)
			os.Exit(1)
		}
		return
	}

	if toolName != "compile" {
		compile.ExecCmd(tool, toolArgs...).Run()
		return
	}

	manager := transform.New()
	transformer, err := manager.Find(importPath)
	if err != nil {
		//ignore
	}

	if transformer != nil {
		transformer.Init(log)
		pkgArgs, err := transformer.Do(args)
		if err != nil {
			log.Error("failed to transform", "err", err)
			os.Exit(1)
		}

		args = pkgArgs
	}

	if err := compile.ExecCmd(tool, toolArgs...).Run(); err != nil {
		if exitErr, ok := err.(*os.PathError); ok {
			log.Error("failed to run tool", "err", exitErr)
		}
		os.Exit(1)
	}
}
