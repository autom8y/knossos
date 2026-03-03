package know

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os/exec"
	"sort"
	"strings"
)

// DeclKind identifies the kind of top-level Go declaration.
type DeclKind string

const (
	DeclFunc  DeclKind = "func"
	DeclType  DeclKind = "type"
	DeclVar   DeclKind = "var"
	DeclConst DeclKind = "const"
)

// ChangeKind classifies how a declaration changed.
type ChangeKind string

const (
	ChangeNew        ChangeKind = "NEW"
	ChangeDeleted    ChangeKind = "DELETED"
	ChangeModified   ChangeKind = "MODIFIED"
	ChangeSigChanged ChangeKind = "SIGNATURE_CHANGED"
)

// DeclChange represents a change to a single top-level declaration.
type DeclChange struct {
	Kind      DeclKind   `json:"kind"`
	Name      string     `json:"name"`
	Receiver  string     `json:"receiver,omitempty"`
	Change    ChangeKind `json:"change"`
	OldSig    string     `json:"old_sig,omitempty"`
	NewSig    string     `json:"new_sig,omitempty"`
	BodyDelta int        `json:"body_delta,omitempty"`
}

// FileDiff represents changes to a single Go source file.
type FileDiff struct {
	Path    string       `json:"path"`
	Changes []DeclChange `json:"changes"`
}

// SemanticDiff is the top-level result for an AST-based diff operation.
type SemanticDiff struct {
	FromHash     string     `json:"from_hash"`
	ToHash       string     `json:"to_hash"`
	Files        []FileDiff `json:"files"`
	NonGoFiles   []string   `json:"non_go_files,omitempty"`
	SkippedFiles []string   `json:"skipped_files,omitempty"`
}

// gitShowFile retrieves file contents at a specific git commit.
// Replaceable in tests to avoid real git invocations.
var gitShowFile = defaultGitShowFile

// GitShowFileFunc returns the current gitShowFile function.
// Used by CLI wiring in other packages to invoke git show through the mockable var.
func GitShowFileFunc() func(string, string) ([]byte, error) {
	return gitShowFile
}

func defaultGitShowFile(commitHash, filePath string) ([]byte, error) {
	out, err := exec.Command("git", "show", commitHash+":"+filePath).Output()
	if err != nil {
		return nil, fmt.Errorf("git show %s:%s: %w", commitHash, filePath, err)
	}
	return out, nil
}

// declaration is the internal representation of a single top-level declaration
// extracted from a parsed Go AST.
type declaration struct {
	kind      DeclKind
	name      string
	receiver  string
	signature string
	bodyLines int
	bodyHash  string
}

// declKey returns the unique identity for deduplication.
// For methods: "receiver.name". For everything else: name.
func (d declaration) declKey() string {
	if d.receiver != "" {
		return d.receiver + "." + d.name
	}
	return d.name
}

// ComputeFileDiff computes declaration-level changes between two versions
// of a Go source file. Returns nil (not an error) when both versions parse
// to identical declaration sets.
//
// oldSource or newSource may be nil:
//   - oldSource == nil: all declarations in newSource are NEW
//   - newSource == nil: all declarations in oldSource are DELETED
//   - both nil: returns nil, nil
func ComputeFileDiff(oldSource, newSource []byte, path string) (*FileDiff, error) {
	if oldSource == nil && newSource == nil {
		return nil, nil
	}

	var oldDecls map[string]declaration
	var newDecls map[string]declaration
	var err error

	if oldSource != nil {
		oldDecls, err = extractDeclarations(oldSource, path)
		if err != nil {
			return nil, fmt.Errorf("parse old %s: %w", path, err)
		}
	}

	if newSource != nil {
		newDecls, err = extractDeclarations(newSource, path)
		if err != nil {
			return nil, fmt.Errorf("parse new %s: %w", path, err)
		}
	}

	if oldDecls == nil {
		oldDecls = make(map[string]declaration)
	}
	if newDecls == nil {
		newDecls = make(map[string]declaration)
	}

	changes := diffDeclarations(oldDecls, newDecls)
	if len(changes) == 0 {
		return nil, nil
	}

	return &FileDiff{
		Path:    path,
		Changes: changes,
	}, nil
}

// FormatSemanticDiff renders a SemanticDiff as a compact markdown string
// suitable for theoros prompt injection.
func FormatSemanticDiff(diff *SemanticDiff) string {
	if diff == nil {
		return ""
	}

	var b strings.Builder

	for _, f := range diff.Files {
		b.WriteString("## ")
		b.WriteString(f.Path)
		b.WriteString("\n")
		for _, c := range f.Changes {
			b.WriteString("- ")
			b.WriteString(string(c.Change))
			b.WriteString(" ")
			b.WriteString(formatChangeEntry(c))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if len(diff.NonGoFiles) > 0 {
		b.WriteString("## Non-Go modified files (no AST diff)\n")
		for _, f := range diff.NonGoFiles {
			b.WriteString("- ")
			b.WriteString(f)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if len(diff.SkippedFiles) > 0 {
		b.WriteString("## Skipped (parse error)\n")
		for _, f := range diff.SkippedFiles {
			b.WriteString("- ")
			b.WriteString(f)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

func formatChangeEntry(c DeclChange) string {
	switch c.Change {
	case ChangeNew:
		return formatDeclSig(c.Kind, c.NewSig)
	case ChangeDeleted:
		return formatDeclSig(c.Kind, c.OldSig)
	case ChangeSigChanged:
		return formatDeclSig(c.Kind, c.NewSig) + "\n  was: " + c.OldSig
	case ChangeModified:
		delta := ""
		if c.BodyDelta != 0 {
			sign := "+"
			val := c.BodyDelta
			if val < 0 {
				sign = ""
			}
			delta = fmt.Sprintf("  [body: %s%d lines]", sign, val)
		}
		return formatDeclSig(c.Kind, c.NewSig) + delta
	default:
		return c.Name
	}
}

func formatDeclSig(kind DeclKind, sig string) string {
	if sig != "" {
		return sig
	}
	return string(kind) + " " + sig
}

// extractDeclarations parses Go source and returns a map of declKey -> declaration.
func extractDeclarations(src []byte, filename string) (map[string]declaration, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	decls := make(map[string]declaration)

	for _, d := range f.Decls {
		switch node := d.(type) {
		case *ast.FuncDecl:
			recv := ""
			if node.Recv != nil && len(node.Recv.List) > 0 {
				recv = extractReceiverName(node.Recv.List[0].Type)
			}
			sig := renderFuncSignature(fset, node)
			bodyLines := countBodyLines(fset, node)
			body := renderFuncBody(fset, node)

			decl := declaration{
				kind:      DeclFunc,
				name:      node.Name.Name,
				receiver:  recv,
				signature: sig,
				bodyLines: bodyLines,
				bodyHash:  body,
			}
			decls[decl.declKey()] = decl

		case *ast.GenDecl:
			if node.Tok == token.IMPORT {
				continue
			}
			kind := tokenToKind(node.Tok)
			for _, spec := range node.Specs {
				decl := extractSpecDeclaration(fset, kind, node, spec)
				decls[decl.declKey()] = decl
			}
		}
	}

	return decls, nil
}

func extractReceiverName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return extractReceiverName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return extractReceiverName(t.X)
	case *ast.IndexListExpr:
		return extractReceiverName(t.X)
	default:
		return ""
	}
}

func tokenToKind(tok token.Token) DeclKind {
	switch tok {
	case token.TYPE:
		return DeclType
	case token.VAR:
		return DeclVar
	case token.CONST:
		return DeclConst
	default:
		return DeclVar
	}
}

func extractSpecDeclaration(fset *token.FileSet, kind DeclKind, genDecl *ast.GenDecl, spec ast.Spec) declaration {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		sig := renderSignature(fset, genDecl, spec)
		return declaration{
			kind:      kind,
			name:      s.Name.Name,
			signature: sig,
		}
	case *ast.ValueSpec:
		name := ""
		if len(s.Names) > 0 {
			name = s.Names[0].Name
		}
		sig := renderSignature(fset, genDecl, spec)
		return declaration{
			kind:      kind,
			name:      name,
			signature: sig,
		}
	default:
		return declaration{kind: kind}
	}
}

// renderFuncSignature renders a FuncDecl without its body using go/printer.
func renderFuncSignature(fset *token.FileSet, fn *ast.FuncDecl) string {
	clone := *fn
	clone.Body = nil

	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.RawFormat}
	if err := cfg.Fprint(&buf, fset, &clone); err != nil {
		return fn.Name.Name
	}
	return normalizeSignature(buf.String())
}

// renderFuncBody renders the body block of a FuncDecl for hash comparison.
func renderFuncBody(fset *token.FileSet, fn *ast.FuncDecl) string {
	if fn.Body == nil {
		return ""
	}
	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.RawFormat}
	if err := cfg.Fprint(&buf, fset, fn.Body); err != nil {
		return ""
	}
	return buf.String()
}

// renderSignature renders a GenDecl spec as a signature string.
func renderSignature(fset *token.FileSet, genDecl *ast.GenDecl, spec ast.Spec) string {
	synth := &ast.GenDecl{
		Tok:   genDecl.Tok,
		Specs: []ast.Spec{spec},
	}

	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.RawFormat}
	if err := cfg.Fprint(&buf, fset, synth); err != nil {
		return ""
	}
	return normalizeSignature(buf.String())
}

func normalizeSignature(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// countBodyLines returns the line count of a FuncDecl body.
// Returns 0 for declarations without a body.
func countBodyLines(fset *token.FileSet, fn *ast.FuncDecl) int {
	if fn.Body == nil {
		return 0
	}
	start := fset.Position(fn.Body.Pos()).Line
	end := fset.Position(fn.Body.End()).Line
	return end - start
}

// diffDeclarations compares old and new declaration maps and produces DeclChange entries.
// Output is sorted by (Kind, Name) for determinism.
func diffDeclarations(old, new map[string]declaration) []DeclChange {
	var changes []DeclChange

	for key, newDecl := range new {
		oldDecl, exists := old[key]
		if !exists {
			changes = append(changes, DeclChange{
				Kind:     newDecl.kind,
				Name:     newDecl.name,
				Receiver: newDecl.receiver,
				Change:   ChangeNew,
				NewSig:   newDecl.signature,
			})
			continue
		}

		if oldDecl.signature != newDecl.signature {
			changes = append(changes, DeclChange{
				Kind:     newDecl.kind,
				Name:     newDecl.name,
				Receiver: newDecl.receiver,
				Change:   ChangeSigChanged,
				OldSig:   oldDecl.signature,
				NewSig:   newDecl.signature,
			})
		} else if newDecl.bodyHash != oldDecl.bodyHash {
			changes = append(changes, DeclChange{
				Kind:      newDecl.kind,
				Name:      newDecl.name,
				Receiver:  newDecl.receiver,
				Change:    ChangeModified,
				NewSig:    newDecl.signature,
				BodyDelta: newDecl.bodyLines - oldDecl.bodyLines,
			})
		}
	}

	for key, oldDecl := range old {
		if _, exists := new[key]; !exists {
			changes = append(changes, DeclChange{
				Kind:     oldDecl.kind,
				Name:     oldDecl.name,
				Receiver: oldDecl.receiver,
				Change:   ChangeDeleted,
				OldSig:   oldDecl.signature,
			})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Kind != changes[j].Kind {
			return changes[i].Kind < changes[j].Kind
		}
		ki := changes[i].Name
		if changes[i].Receiver != "" {
			ki = changes[i].Receiver + "." + changes[i].Name
		}
		kj := changes[j].Name
		if changes[j].Receiver != "" {
			kj = changes[j].Receiver + "." + changes[j].Name
		}
		return ki < kj
	})

	return changes
}
