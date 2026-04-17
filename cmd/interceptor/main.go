package main

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/davewhit3/compile-interceptor/compile"
	"github.com/davewhit3/compile-interceptor/outgoing"
	"github.com/davewhit3/compile-interceptor/transform"

	_ "net/http"

	_ "github.com/valkey-io/valkey-go"
)

func main() {
	compile.InjectedDepsHash = outgoing.SourceHash

	level := slog.LevelInfo
	if v := os.Getenv("INTERCEPTOR_LOG_LEVEL"); v != "" {
		_ = level.UnmarshalText([]byte(v))
	}
	slog.SetLogLoggerLevel(level)
	log := slog.Default()

	args := os.Args[compile.ExecCmdArgsOffset:]
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: compile-interceptor <compiler> [args...]\n")
		os.Exit(1)
	}

	importPath := os.Getenv("TOOLEXEC_IMPORTPATH")
	workDir := compile.DeriveWorkDir(args)

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
		pkgArgs, err := transformer.Do(toolArgs)
		if err != nil {
			log.Error("failed to transform", "err", err)
			os.Exit(1)
		}

		toolArgs = pkgArgs
	}

	if err := compile.ExecCmd(tool, toolArgs...).Run(); err != nil {
		if exitErr, ok := err.(*os.PathError); ok {
			log.Error("failed to run tool", "err", exitErr)
		}
		os.Exit(1)
	}

	if importPath == transform.OutgoingPkgPath {
		outputPath := compile.ExtractOutputPath(toolArgs)
		if workDir != "" && outputPath != "" {
			if err := compile.SavePkgArchivePath(workDir, importPath, outputPath); err != nil {
				log.Error("failed to save outgoing archive path", "err", err)
			}
		}
	}
}
