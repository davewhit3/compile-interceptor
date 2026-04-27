package transform

import (
	"fmt"
	"go/token"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/davewhit3/compile-interceptor/compile"
)

const OutgoingPkgPath = "github.com/davewhit3/compile-interceptor/outgoing"
const outgoingImport = `"` + OutgoingPkgPath + `"`


var transforms []Transform = make([]Transform, 0)

type Transform interface {
	Init(logger *slog.Logger)
	Do(args []string) ([]string, error)
	Support(importPath string) bool
}

type Manager struct {
	transforms []Transform
}

func Register(transform Transform) {
	transforms = append(transforms, transform)
}

func New() *Manager {
	return &Manager{}
}

func (m *Manager) Find(importPath string) (Transform, error) {
	for _, transform := range transforms {
		if transform.Support(importPath) {
			return transform, nil
		}
	}

	return nil, fmt.Errorf("transformer for import path %s not found", importPath)
}

// transformer is a struct that contains the imports and code to be transformed
type transformer struct {
	workDir           string
	logger            *slog.Logger
	SourcePackage     string
	SourceCode        string
	SourceFile        string
	SourceFilePattern *regexp.Regexp
	TemplateCode      string
	TargetFunc        string
	Imports           []string
	// InjectedPkg overrides the default outgoing package that is injected into
	// the transformed file. When empty, OutgoingPkgPath is used.
	InjectedPkg string
	Code        *dst.File
	Template    *dst.File
}

// injectedPkgPath returns the import path of the package that will be injected.
func (t *transformer) injectedPkgPath() string {
	if t.InjectedPkg != "" {
		return t.InjectedPkg
	}
	return OutgoingPkgPath
}

// injectedPkgImport returns the quoted import string for the injected package.
func (t *transformer) injectedPkgImport() string {
	if t.InjectedPkg != "" {
		return `"` + t.InjectedPkg + `"`
	}
	return outgoingImport
}

func (t *transformer) Init(logger *slog.Logger) {
	t.logger = logger
}

func parseCode(code string) (*dst.File, error) {
	return decorator.Parse(code)
}

func parseFile(file string) (string, error) {
	rawCode, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", file, err)
	}

	return string(rawCode), nil
}

func (t *transformer) LoadCode(file string) error {
	rawCode, err := parseFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", file, err)
	}

	code, err := parseCode(rawCode)
	if err != nil {
		return fmt.Errorf("failed to parse original file %s: %w", file, err)
	}

	t.Code = code
	return nil
}

func (t *transformer) LoadTemplate() error {
	code, err := parseCode(t.TemplateCode)
	if err != nil {
		return fmt.Errorf("failed to parse template code: %w", err)
	}

	t.Template = code
	return nil
}

func (t *transformer) Transform() error {
	var modFn *dst.FuncDecl

	for _, dec := range t.Template.Decls {
		if d, ok := dec.(*dst.FuncDecl); ok {
			if d.Name.Name == t.TargetFunc {
				modFn = d
				break
			}
		}
	}

	for _, dec := range t.Code.Decls {
		if d, ok := dec.(*dst.FuncDecl); ok {
			if d.Name.Name == t.TargetFunc {
				d.Body = modFn.Body
				break
			}
		}
	}

	t.AddImports()

	return nil
}

func (t *transformer) SaveModFile(file string) (string, error) {
	replacer := strings.NewReplacer("/", "_", ".", "_")
	mf := t.workDir + "/" + replacer.Replace(t.SourcePackage) + "_" + strings.TrimRight(filepath.Base(file), ".go") + "_" + strconv.FormatInt(time.Now().UnixNano(), 10) + "_mod.go"

	t.logger.Debug("writing mod file", "file", mf)

	tf, err := os.OpenFile(mf, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file %s: %w", mf, err)
	}

	err = decorator.Fprint(tf, t.Code)
	if err != nil {
		return "", fmt.Errorf("failed to print code to temp file %s: %w", mf, err)
	}

	if err := tf.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file %s: %w", mf, err)
	}

	return tf.Name(), nil
}

func (t *transformer) AddImports() {
	for _, imp := range append(t.Imports, t.injectedPkgImport()) {
		newImport := &dst.ImportSpec{
			Path: &dst.BasicLit{
				Kind:  token.STRING,
				Value: imp,
			},
		}
		newImport.Decs.Before = dst.NewLine
		newImport.Decs.After = dst.NewLine

		t.Code.Imports = append(t.Code.Imports, newImport)

		inserted := false
		for _, decl := range t.Code.Decls {
			genDecl, ok := decl.(*dst.GenDecl)
			if ok && genDecl.Tok == token.IMPORT {
				genDecl.Specs = append(genDecl.Specs, newImport)
				inserted = true
				break
			}
		}

		if !inserted {
			newDecl := &dst.GenDecl{
				Tok:    token.IMPORT,
				Specs:  []dst.Spec{newImport},
				Lparen: true,
				Rparen: true,
			}

			t.Code.Decls = append([]dst.Decl{newDecl}, t.Code.Decls...)
		}
	}
}

func (t *transformer) Do(args []string) ([]string, error) {
	filesToCompile, idx, _ := compile.ExtractFilesFromPack(args)
	for i, file := range filesToCompile {
		matched := false
		if t.SourceFilePattern != nil {
			matched = t.SourceFilePattern.MatchString(file)
		} else {
			matched = strings.HasSuffix(file, t.SourceFile)
		}

		if matched {
			t.logger.Debug("processing file", "file", file, "sourceFile", t.SourceFile)
			t.workDir = compile.DeriveWorkDir(args)

			// Go compiles sibling dependencies in parallel, so the injected
			// package may not be compiled by the outer build yet. Build it
			// ourselves (idempotent, flock-guarded) so the injected import
			// resolves during this compile pass.
			archivePath, err := compile.EnsurePkgArchive(t.workDir, t.injectedPkgPath())
			if err != nil {
				return nil, fmt.Errorf("ensuring outgoing archive: %w", err)
			}
			t.logger.Debug("outgoing archive ready", "archive", archivePath)

			t.logger.Debug("loading code for file", "file", file)
			if err := t.LoadCode(file); err != nil {
				return nil, fmt.Errorf("failed to load code for file %s: %w", file, err)
			}

			t.logger.Debug("loading template")
			if err := t.LoadTemplate(); err != nil {
				return nil, fmt.Errorf("failed to load template: %w", err)
			}

			t.logger.Debug("transforming file", "file", file)
			if err := t.Transform(); err != nil {
				return nil, fmt.Errorf("failed to transform file %s: %w", file, err)
			}

			modFile, err := t.SaveModFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to save mod file for file %s: %w", file, err)
			}

			t.logger.Info("transformed", "file", file, "modFile", modFile)

			args[idx+i-1] = modFile

			if err := t.injectPkgDep(args, t.injectedPkgPath()); err != nil {
				return nil, fmt.Errorf("failed to inject %s dependency: %w", t.injectedPkgPath(), err)
			}

			return args, nil
		}
	}

	return nil, fmt.Errorf("file %s not found", t.SourceFile)
}

func (t *transformer) injectPkgDep(args []string, pkgPath string) error {
	archivePath, err := compile.LoadPkgArchivePath(t.workDir, pkgPath)
	if err != nil {
		return fmt.Errorf("loading archive path for %s: %w", pkgPath, err)
	}

	importcfgPath := compile.ExtractImportCfgPath(args)
	if importcfgPath == "" {
		return fmt.Errorf("importcfg path not found in args")
	}

	t.logger.Debug("injecting pkg into importcfg", "pkg", pkgPath, "archive", archivePath, "importcfg", importcfgPath)

	if err := compile.AddPackage(importcfgPath, pkgPath, archivePath); err != nil {
		return fmt.Errorf("adding %s to importcfg: %w", pkgPath, err)
	}

	return nil
}

// injectOutgoingDep is kept for backward compatibility.
func (t *transformer) injectOutgoingDep(args []string) error {
	return t.injectPkgDep(args, OutgoingPkgPath)
}

func (t *transformer) Support(importPath string) bool {
	return importPath == t.SourcePackage
}
