package know

import (
	"fmt"
	"strings"
	"testing"
)

// --- Group 1: extractDeclarations ---

func TestExtractDeclarations(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		wantKeys  []string
		wantKinds map[string]DeclKind
		wantRecv  map[string]string
		wantErr   bool
	}{
		{
			name:     "EmptyFile",
			src:      "package foo\n",
			wantKeys: nil,
		},
		{
			name:    "EmptySource",
			src:     "",
			wantErr: true,
		},
		{
			name:      "SingleFunc",
			src:       "package foo\n\nfunc Hello() {}\n",
			wantKeys:  []string{"Hello"},
			wantKinds: map[string]DeclKind{"Hello": DeclFunc},
		},
		{
			name:      "FuncWithParams",
			src:       "package foo\n\nfunc Add(a, b int) int { return a + b }\n",
			wantKeys:  []string{"Add"},
			wantKinds: map[string]DeclKind{"Add": DeclFunc},
		},
		{
			name:      "Method",
			src:       "package foo\n\ntype Meta struct{}\n\nfunc (m *Meta) Domain() string { return \"\" }\n",
			wantKeys:  []string{"Meta", "Meta.Domain"},
			wantKinds: map[string]DeclKind{"Meta.Domain": DeclFunc, "Meta": DeclType},
			wantRecv:  map[string]string{"Meta.Domain": "Meta"},
		},
		{
			name:      "TypeStruct",
			src:       "package foo\n\ntype Config struct { Name string }\n",
			wantKeys:  []string{"Config"},
			wantKinds: map[string]DeclKind{"Config": DeclType},
		},
		{
			name:      "TypeInterface",
			src:       "package foo\n\ntype Reader interface { Read([]byte) (int, error) }\n",
			wantKeys:  []string{"Reader"},
			wantKinds: map[string]DeclKind{"Reader": DeclType},
		},
		{
			name:      "VarDecl",
			src:       "package foo\n\nvar DefaultTimeout = 30\n",
			wantKeys:  []string{"DefaultTimeout"},
			wantKinds: map[string]DeclKind{"DefaultTimeout": DeclVar},
		},
		{
			name:      "ConstDecl",
			src:       "package foo\n\nconst MaxRetries = 3\n",
			wantKeys:  []string{"MaxRetries"},
			wantKinds: map[string]DeclKind{"MaxRetries": DeclConst},
		},
		{
			name:      "GroupedConst",
			src:       "package foo\n\nconst (\n\tA = 1\n\tB = 2\n\tC = 3\n)\n",
			wantKeys:  []string{"A", "B", "C"},
			wantKinds: map[string]DeclKind{"A": DeclConst, "B": DeclConst, "C": DeclConst},
		},
		{
			name:     "ImportIgnored",
			src:      "package foo\n\nimport \"fmt\"\n\nvar _ = fmt.Println\n",
			wantKeys: []string{"_"},
		},
		{
			name:      "Generics",
			src:       "package foo\n\nfunc Map[T any](s []T, f func(T) T) []T { return nil }\n",
			wantKeys:  []string{"Map"},
			wantKinds: map[string]DeclKind{"Map": DeclFunc},
		},
		{
			name:      "GenericType",
			src:       "package foo\n\ntype Set[T comparable] struct { m map[T]bool }\n",
			wantKeys:  []string{"Set"},
			wantKinds: map[string]DeclKind{"Set": DeclType},
		},
		{
			name:    "ParseError",
			src:     "package foo\n\nfunc { broken\n",
			wantErr: true,
		},
		{
			name: "MultipleReceivers",
			src: `package foo

type T struct{}
func (t *T) Foo() {}
func (t *T) Bar() {}
`,
			wantKeys:  []string{"T", "T.Foo", "T.Bar"},
			wantKinds: map[string]DeclKind{"T": DeclType, "T.Foo": DeclFunc, "T.Bar": DeclFunc},
			wantRecv:  map[string]string{"T.Foo": "T", "T.Bar": "T"},
		},
		{
			name:      "GroupedVar",
			src:       "package foo\n\nvar (\n\tX = 1\n\tY = 2\n)\n",
			wantKeys:  []string{"X", "Y"},
			wantKinds: map[string]DeclKind{"X": DeclVar, "Y": DeclVar},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decls, err := extractDeclarations([]byte(tt.src), "test.go")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantKeys == nil {
				if len(decls) != 0 {
					t.Errorf("expected empty map, got %d entries: %v", len(decls), keys(decls))
				}
				return
			}

			for _, key := range tt.wantKeys {
				if _, ok := decls[key]; !ok {
					t.Errorf("missing expected key %q, have: %v", key, keys(decls))
				}
			}
			if len(decls) != len(tt.wantKeys) {
				t.Errorf("got %d declarations, want %d. keys: %v", len(decls), len(tt.wantKeys), keys(decls))
			}

			for key, wantKind := range tt.wantKinds {
				if d, ok := decls[key]; ok {
					if d.kind != wantKind {
						t.Errorf("decl %q kind = %q, want %q", key, d.kind, wantKind)
					}
				}
			}

			for key, wantRecv := range tt.wantRecv {
				if d, ok := decls[key]; ok {
					if d.receiver != wantRecv {
						t.Errorf("decl %q receiver = %q, want %q", key, d.receiver, wantRecv)
					}
				}
			}
		})
	}
}

func TestExtractDeclarations_SignatureContent(t *testing.T) {
	src := "package foo\n\nfunc Add(a, b int) int { return a + b }\n"
	decls, err := extractDeclarations([]byte(src), "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := decls["Add"]
	if !strings.Contains(d.signature, "a, b int") {
		t.Errorf("signature should contain params, got: %q", d.signature)
	}
	if !strings.Contains(d.signature, "int") {
		t.Errorf("signature should contain return type, got: %q", d.signature)
	}
}

func TestExtractDeclarations_GenericSignature(t *testing.T) {
	src := "package foo\n\nfunc Map[T any](s []T, f func(T) T) []T { return nil }\n"
	decls, err := extractDeclarations([]byte(src), "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := decls["Map"]
	if !strings.Contains(d.signature, "[T any]") {
		t.Errorf("generic signature should contain type params, got: %q", d.signature)
	}
}

func TestExtractDeclarations_BodyLines(t *testing.T) {
	src := `package foo

func Multi() {
	a := 1
	b := 2
	c := a + b
	_ = c
}
`
	decls, err := extractDeclarations([]byte(src), "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := decls["Multi"]
	if d.bodyLines == 0 {
		t.Error("bodyLines should be non-zero for a function with a body")
	}
}

// --- Group 2: diffDeclarations ---

func TestDiffDeclarations(t *testing.T) {
	tests := []struct {
		name        string
		old         map[string]declaration
		new         map[string]declaration
		wantChanges []struct {
			name   string
			change ChangeKind
		}
	}{
		{
			name: "NoChanges",
			old:  map[string]declaration{"A": {kind: DeclFunc, name: "A", signature: "func A()", bodyLines: 5, bodyHash: "x"}},
			new:  map[string]declaration{"A": {kind: DeclFunc, name: "A", signature: "func A()", bodyLines: 5, bodyHash: "x"}},
		},
		{
			name: "NewFunc",
			old:  map[string]declaration{},
			new:  map[string]declaration{"A": {kind: DeclFunc, name: "A", signature: "func A()"}},
			wantChanges: []struct {
				name   string
				change ChangeKind
			}{{"A", ChangeNew}},
		},
		{
			name: "DeletedFunc",
			old:  map[string]declaration{"A": {kind: DeclFunc, name: "A", signature: "func A()"}},
			new:  map[string]declaration{},
			wantChanges: []struct {
				name   string
				change ChangeKind
			}{{"A", ChangeDeleted}},
		},
		{
			name: "SignatureChanged",
			old:  map[string]declaration{"A": {kind: DeclFunc, name: "A", signature: "func A(x int)"}},
			new:  map[string]declaration{"A": {kind: DeclFunc, name: "A", signature: "func A(x string)"}},
			wantChanges: []struct {
				name   string
				change ChangeKind
			}{{"A", ChangeSigChanged}},
		},
		{
			name: "BodyModified",
			old:  map[string]declaration{"A": {kind: DeclFunc, name: "A", signature: "func A()", bodyLines: 10, bodyHash: "old"}},
			new:  map[string]declaration{"A": {kind: DeclFunc, name: "A", signature: "func A()", bodyLines: 15, bodyHash: "new"}},
			wantChanges: []struct {
				name   string
				change ChangeKind
			}{{"A", ChangeModified}},
		},
		{
			name: "BodyModifiedSameLineCount",
			old:  map[string]declaration{"A": {kind: DeclFunc, name: "A", signature: "func A()", bodyLines: 10, bodyHash: "old"}},
			new:  map[string]declaration{"A": {kind: DeclFunc, name: "A", signature: "func A()", bodyLines: 10, bodyHash: "new"}},
			wantChanges: []struct {
				name   string
				change ChangeKind
			}{{"A", ChangeModified}},
		},
		{
			name: "MixedChanges",
			old: map[string]declaration{
				"A": {kind: DeclFunc, name: "A", signature: "func A()"},
				"B": {kind: DeclFunc, name: "B", signature: "func B()", bodyLines: 5, bodyHash: "old"},
				"C": {kind: DeclFunc, name: "C", signature: "func C()"},
			},
			new: map[string]declaration{
				"B": {kind: DeclFunc, name: "B", signature: "func B()", bodyLines: 8, bodyHash: "new"},
				"D": {kind: DeclFunc, name: "D", signature: "func D()"},
			},
			wantChanges: []struct {
				name   string
				change ChangeKind
			}{
				{"A", ChangeDeleted},
				{"B", ChangeModified},
				{"C", ChangeDeleted},
				{"D", ChangeNew},
			},
		},
		{
			name: "ReceiverAdded",
			old:  map[string]declaration{"Foo": {kind: DeclFunc, name: "Foo", signature: "func Foo()"}},
			new:  map[string]declaration{"T.Foo": {kind: DeclFunc, name: "Foo", receiver: "T", signature: "func (t T) Foo()"}},
			wantChanges: []struct {
				name   string
				change ChangeKind
			}{
				{"Foo", ChangeDeleted},
				{"Foo", ChangeNew}, // T.Foo sorts as func T.Foo
			},
		},
		{
			name: "TypeChanged",
			old:  map[string]declaration{"Config": {kind: DeclType, name: "Config", signature: "type Config struct{}"}},
			new:  map[string]declaration{"Config": {kind: DeclType, name: "Config", signature: "type Config interface{}"}},
			wantChanges: []struct {
				name   string
				change ChangeKind
			}{{"Config", ChangeSigChanged}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := diffDeclarations(tt.old, tt.new)

			if tt.wantChanges == nil {
				if len(changes) != 0 {
					t.Errorf("expected no changes, got %d: %v", len(changes), changes)
				}
				return
			}

			if len(changes) != len(tt.wantChanges) {
				t.Fatalf("got %d changes, want %d.\nGot: %v", len(changes), len(tt.wantChanges), changes)
			}

			for i, want := range tt.wantChanges {
				if changes[i].Name != want.name {
					t.Errorf("change[%d].Name = %q, want %q", i, changes[i].Name, want.name)
				}
				if changes[i].Change != want.change {
					t.Errorf("change[%d].Change = %q, want %q", i, changes[i].Change, want.change)
				}
			}
		})
	}
}

func TestDiffDeclarations_BodyDelta(t *testing.T) {
	old := map[string]declaration{
		"A": {kind: DeclFunc, name: "A", signature: "func A()", bodyLines: 10, bodyHash: "old"},
	}
	new := map[string]declaration{
		"A": {kind: DeclFunc, name: "A", signature: "func A()", bodyLines: 15, bodyHash: "new"},
	}
	changes := diffDeclarations(old, new)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].BodyDelta != 5 {
		t.Errorf("BodyDelta = %d, want 5", changes[0].BodyDelta)
	}
}

func TestDiffDeclarations_SortOrder(t *testing.T) {
	old := map[string]declaration{}
	new := map[string]declaration{
		"Z":     {kind: DeclVar, name: "Z", signature: "var Z int"},
		"A":     {kind: DeclFunc, name: "A", signature: "func A()"},
		"M":     {kind: DeclType, name: "M", signature: "type M struct{}"},
		"B":     {kind: DeclFunc, name: "B", signature: "func B()"},
		"Alpha": {kind: DeclConst, name: "Alpha", signature: "const Alpha = 1"},
	}
	changes := diffDeclarations(old, new)

	if len(changes) != 5 {
		t.Fatalf("expected 5 changes, got %d", len(changes))
	}

	// Expected sort: const Alpha, func A, func B, type M, var Z
	expected := []struct {
		kind DeclKind
		name string
	}{
		{DeclConst, "Alpha"},
		{DeclFunc, "A"},
		{DeclFunc, "B"},
		{DeclType, "M"},
		{DeclVar, "Z"},
	}
	for i, want := range expected {
		if changes[i].Kind != want.kind || changes[i].Name != want.name {
			t.Errorf("changes[%d] = {%s, %s}, want {%s, %s}",
				i, changes[i].Kind, changes[i].Name, want.kind, want.name)
		}
	}
}

// --- Group 3: ComputeFileDiff ---

func TestComputeFileDiff_BothNil(t *testing.T) {
	diff, err := ComputeFileDiff(nil, nil, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff != nil {
		t.Errorf("both nil should return nil diff, got: %v", diff)
	}
}

func TestComputeFileDiff_OldNil(t *testing.T) {
	newSrc := []byte("package foo\n\nfunc Hello() {}\nfunc World() {}\n")
	diff, err := ComputeFileDiff(nil, newSrc, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff == nil {
		t.Fatal("expected non-nil diff when old is nil")
	}
	for _, c := range diff.Changes {
		if c.Change != ChangeNew {
			t.Errorf("all changes should be NEW when old is nil, got %s for %s", c.Change, c.Name)
		}
	}
}

func TestComputeFileDiff_NewNil(t *testing.T) {
	oldSrc := []byte("package foo\n\nfunc Hello() {}\nfunc World() {}\n")
	diff, err := ComputeFileDiff(oldSrc, nil, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff == nil {
		t.Fatal("expected non-nil diff when new is nil")
	}
	for _, c := range diff.Changes {
		if c.Change != ChangeDeleted {
			t.Errorf("all changes should be DELETED when new is nil, got %s for %s", c.Change, c.Name)
		}
	}
}

func TestComputeFileDiff_IdenticalSource(t *testing.T) {
	src := []byte("package foo\n\nfunc Hello() { println(\"hello\") }\n")
	diff, err := ComputeFileDiff(src, src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff != nil {
		t.Errorf("identical source should return nil diff, got: %v", diff)
	}
}

func TestComputeFileDiff_RealWorldChange(t *testing.T) {
	oldSrc := []byte(`package foo

func Existing() int {
	return 1
}

func ToDelete() {}
`)
	newSrc := []byte(`package foo

func Existing() int {
	x := 1
	y := 2
	return x + y
}

func NewFunc() string {
	return "new"
}
`)
	diff, err := ComputeFileDiff(oldSrc, newSrc, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff == nil {
		t.Fatal("expected non-nil diff for changed source")
	}

	changeMap := make(map[string]ChangeKind)
	for _, c := range diff.Changes {
		changeMap[c.Name] = c.Change
	}

	if changeMap["Existing"] != ChangeModified {
		t.Errorf("Existing should be MODIFIED, got %s", changeMap["Existing"])
	}
	if changeMap["ToDelete"] != ChangeDeleted {
		t.Errorf("ToDelete should be DELETED, got %s", changeMap["ToDelete"])
	}
	if changeMap["NewFunc"] != ChangeNew {
		t.Errorf("NewFunc should be NEW, got %s", changeMap["NewFunc"])
	}
}

func TestComputeFileDiff_ParseError(t *testing.T) {
	oldSrc := []byte("package foo\n\nfunc Hello() {}\n")
	newSrc := []byte("package foo\n\nfunc { broken\n")
	_, err := ComputeFileDiff(oldSrc, newSrc, "test.go")
	if err == nil {
		t.Fatal("expected error for unparseable new source")
	}
}

// --- Group 4: FormatSemanticDiff ---

func TestFormatSemanticDiff_EmptyDiff(t *testing.T) {
	diff := &SemanticDiff{
		FromHash: "abc",
		ToHash:   "def",
	}
	result := FormatSemanticDiff(diff)
	if result != "" {
		t.Errorf("empty diff should produce empty output, got: %q", result)
	}
}

func TestFormatSemanticDiff_NilDiff(t *testing.T) {
	result := FormatSemanticDiff(nil)
	if result != "" {
		t.Errorf("nil diff should produce empty output, got: %q", result)
	}
}

func TestFormatSemanticDiff_SingleFile(t *testing.T) {
	diff := &SemanticDiff{
		FromHash: "abc",
		ToHash:   "def",
		Files: []FileDiff{
			{
				Path: "internal/know/manifest.go",
				Changes: []DeclChange{
					{Kind: DeclFunc, Name: "Hello", Change: ChangeNew, NewSig: "func Hello()"},
					{Kind: DeclFunc, Name: "World", Change: ChangeDeleted, OldSig: "func World()"},
				},
			},
		},
	}
	result := FormatSemanticDiff(diff)
	if !strings.Contains(result, "## internal/know/manifest.go") {
		t.Errorf("output should contain file header, got: %q", result)
	}
	if !strings.Contains(result, "NEW") {
		t.Errorf("output should contain NEW, got: %q", result)
	}
	if !strings.Contains(result, "DELETED") {
		t.Errorf("output should contain DELETED, got: %q", result)
	}
}

func TestFormatSemanticDiff_NonGoFiles(t *testing.T) {
	diff := &SemanticDiff{
		FromHash:   "abc",
		ToHash:     "def",
		NonGoFiles: []string{"go.mod", "docs/README.md"},
	}
	result := FormatSemanticDiff(diff)
	if !strings.Contains(result, "Non-Go modified files") {
		t.Errorf("output should contain non-Go section, got: %q", result)
	}
	if !strings.Contains(result, "go.mod") {
		t.Errorf("output should list go.mod, got: %q", result)
	}
}

func TestFormatSemanticDiff_SkippedFiles(t *testing.T) {
	diff := &SemanticDiff{
		FromHash:     "abc",
		ToHash:       "def",
		SkippedFiles: []string{"broken.go"},
	}
	result := FormatSemanticDiff(diff)
	if !strings.Contains(result, "Skipped") {
		t.Errorf("output should contain skipped section, got: %q", result)
	}
	if !strings.Contains(result, "broken.go") {
		t.Errorf("output should list broken.go, got: %q", result)
	}
}

func TestFormatSemanticDiff_SortedOutput(t *testing.T) {
	// Changes are already sorted by diffDeclarations before reaching FormatSemanticDiff.
	// FormatSemanticDiff preserves the order it receives.
	diff := &SemanticDiff{
		FromHash: "abc",
		ToHash:   "def",
		Files: []FileDiff{
			{
				Path: "test.go",
				Changes: []DeclChange{
					{Kind: DeclConst, Name: "A", Change: ChangeNew, NewSig: "const A = 1"},
					{Kind: DeclFunc, Name: "B", Change: ChangeNew, NewSig: "func B()"},
					{Kind: DeclVar, Name: "Z", Change: ChangeNew, NewSig: "var Z int"},
				},
			},
		},
	}
	result := FormatSemanticDiff(diff)
	idxA := strings.Index(result, "const A")
	idxB := strings.Index(result, "func B")
	idxZ := strings.Index(result, "var Z")
	if idxA < 0 || idxB < 0 || idxZ < 0 {
		t.Fatalf("missing expected entries in output: %q", result)
	}
	if idxA > idxB || idxB > idxZ {
		t.Errorf("changes should be in sorted order (const, func, var), got: %q", result)
	}
}

func TestFormatSemanticDiff_ModifiedWithBodyDelta(t *testing.T) {
	diff := &SemanticDiff{
		FromHash: "abc",
		ToHash:   "def",
		Files: []FileDiff{
			{
				Path: "test.go",
				Changes: []DeclChange{
					{Kind: DeclFunc, Name: "Foo", Change: ChangeModified, NewSig: "func Foo()", BodyDelta: 5},
				},
			},
		},
	}
	result := FormatSemanticDiff(diff)
	if !strings.Contains(result, "[body: +5 lines]") {
		t.Errorf("output should contain body delta, got: %q", result)
	}
}

func TestFormatSemanticDiff_SigChanged(t *testing.T) {
	diff := &SemanticDiff{
		FromHash: "abc",
		ToHash:   "def",
		Files: []FileDiff{
			{
				Path: "test.go",
				Changes: []DeclChange{
					{Kind: DeclFunc, Name: "Foo", Change: ChangeSigChanged,
						OldSig: "func Foo(x int)", NewSig: "func Foo(x string)"},
				},
			},
		},
	}
	result := FormatSemanticDiff(diff)
	if !strings.Contains(result, "SIGNATURE_CHANGED") {
		t.Errorf("output should contain SIGNATURE_CHANGED, got: %q", result)
	}
	if !strings.Contains(result, "was:") {
		t.Errorf("output should contain 'was:' for old signature, got: %q", result)
	}
}

// --- Group 5: Edge cases ---

func TestComputeFileDiff_EmptyFile(t *testing.T) {
	oldSrc := []byte("package foo\n\nfunc Hello() {}\n")
	newSrc := []byte("package foo\n")
	diff, err := ComputeFileDiff(oldSrc, newSrc, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff == nil {
		t.Fatal("expected non-nil diff when going from declarations to empty file")
	}
	for _, c := range diff.Changes {
		if c.Change != ChangeDeleted {
			t.Errorf("expected DELETED for removed declaration, got %s", c.Change)
		}
	}
}

func TestComputeFileDiff_MethodReceiverChange(t *testing.T) {
	oldSrc := []byte(`package foo

type T struct{}

func (t T) Method() {}
`)
	newSrc := []byte(`package foo

type T struct{}

func (t *T) Method() {}
`)
	diff, err := ComputeFileDiff(oldSrc, newSrc, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Both have receiver "T", key is "T.Method", but signature changed (value vs pointer receiver)
	if diff == nil {
		t.Fatal("expected non-nil diff for receiver type change")
	}
	found := false
	for _, c := range diff.Changes {
		if c.Name == "Method" && c.Change == ChangeSigChanged {
			found = true
		}
	}
	if !found {
		t.Errorf("expected SIGNATURE_CHANGED for Method, got changes: %v", diff.Changes)
	}
}

// keys returns the keys of a declaration map for diagnostics.
func keys(m map[string]declaration) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// =============================================================================
// Adversarial Tests -- QA Adversary validation
// =============================================================================

// TestAdversarial_WhitespaceOnlyBodyChange verifies that reformatting a function
// body (same logic, different whitespace) is NOT reported as MODIFIED when the
// AST-printed body normalizes identically. The bodyHash is produced by go/printer
// which normalizes formatting, so pure whitespace changes should be invisible.
func TestAdversarial_WhitespaceOnlyBodyChange(t *testing.T) {
	oldSrc := []byte(`package foo

func Hello() string {
	x := "hello"
	return x
}
`)
	// Same logic, different whitespace/formatting
	newSrc := []byte(`package foo

func Hello() string {
	x :=    "hello"
	return    x
}
`)
	diff, err := ComputeFileDiff(oldSrc, newSrc, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// go/printer normalizes whitespace, so the bodyHash should be identical.
	// If the implementation uses raw source text instead of AST-printed body,
	// this would be a false positive MODIFIED.
	if diff != nil {
		t.Errorf("whitespace-only body change should not produce a diff, got changes: %v", diff.Changes)
	}
}

// TestAdversarial_CommentOnlyBodyChange verifies that adding/removing comments
// inside a function body does NOT produce a false MODIFIED, since go/printer
// with ParseComments may or may not include comments in the body hash.
func TestAdversarial_CommentOnlyBodyChange(t *testing.T) {
	oldSrc := []byte(`package foo

func Hello() string {
	return "hello"
}
`)
	newSrc := []byte(`package foo

func Hello() string {
	// this is a new comment
	return "hello"
}
`)
	diff, err := ComputeFileDiff(oldSrc, newSrc, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Note: whether comments trigger MODIFIED depends on whether go/printer
	// includes comments in the rendered body. This test documents the behavior
	// rather than prescribing it. The key property is: no panic, no error.
	// If comments DO trigger MODIFIED, that is acceptable (better to over-report
	// than under-report). Just document the behavior.
	if diff != nil {
		t.Logf("INFO: comment-only body change IS detected as MODIFIED (bodyHash includes comments)")
		for _, c := range diff.Changes {
			if c.Name == "Hello" && c.Change != ChangeModified {
				t.Errorf("expected MODIFIED for Hello if detected, got %s", c.Change)
			}
		}
	} else {
		t.Logf("INFO: comment-only body change is NOT detected (bodyHash excludes comments)")
	}
}

// TestAdversarial_ClosuresNotExtracted verifies that anonymous functions and
// closures defined inside a function body are NOT extracted as top-level
// declarations. Only the enclosing function should appear.
func TestAdversarial_ClosuresNotExtracted(t *testing.T) {
	src := []byte(`package foo

func Outer() {
	inner := func() int {
		return 42
	}
	_ = inner

	go func() {
		println("goroutine")
	}()

	defer func() {
		recover()
	}()
}
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decls) != 1 {
		t.Errorf("expected exactly 1 declaration (Outer), got %d: %v", len(decls), keys(decls))
	}
	if _, ok := decls["Outer"]; !ok {
		t.Errorf("expected declaration key 'Outer', got: %v", keys(decls))
	}
}

// TestAdversarial_MultipleInitFunctions tests that Go allows multiple init()
// functions in a single file. Since the declKey for all of them is "init",
// this creates a key collision. The implementation should handle this gracefully
// (last one wins is acceptable, or all are tracked).
func TestAdversarial_MultipleInitFunctions(t *testing.T) {
	src := []byte(`package foo

func init() {
	println("first")
}

func init() {
	println("second")
}

func init() {
	println("third")
}
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Key collision: all three init() functions map to key "init".
	// The map can only hold one entry. This is a known limitation.
	// The important thing is: no panic, no error.
	initDecl, ok := decls["init"]
	if !ok {
		t.Fatal("expected at least one 'init' declaration in the map")
	}
	if initDecl.kind != DeclFunc {
		t.Errorf("init declaration should be DeclFunc, got %s", initDecl.kind)
	}
	t.Logf("INFO: %d init() functions in source, %d in declaration map (key collision expected)", 3, 1)
}

// TestAdversarial_MultipleInitDiffing tests that when init() functions change
// between versions, the diff handles the key collision gracefully.
func TestAdversarial_MultipleInitDiffing(t *testing.T) {
	oldSrc := []byte(`package foo

func init() {
	println("old")
}

func init() {
	println("also old")
}
`)
	newSrc := []byte(`package foo

func init() {
	println("new")
}

func init() {
	println("also new")
}

func init() {
	println("third new")
}
`)
	// Should not panic or error, even with init() key collisions
	diff, err := ComputeFileDiff(oldSrc, newSrc, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The exact behavior (MODIFIED or no change) depends on which init()
	// wins the key collision. The critical property is: no crash.
	t.Logf("INFO: diff result for multiple init() changes: %v", diff)
}

// TestAdversarial_BlankIdentifierReceiver tests the edge case where a method
// has a blank identifier receiver: func (_ T) Method().
func TestAdversarial_BlankIdentifierReceiver(t *testing.T) {
	src := []byte(`package foo

type T struct{}

func (_ T) Method() string { return "" }
func (_ *T) PtrMethod() string { return "" }
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The receiver name should be "T" (from the type, not the blank identifier)
	if d, ok := decls["T.Method"]; ok {
		if d.receiver != "T" {
			t.Errorf("blank identifier receiver: got receiver %q, want %q", d.receiver, "T")
		}
	} else {
		t.Errorf("expected key 'T.Method', got: %v", keys(decls))
	}
	if d, ok := decls["T.PtrMethod"]; ok {
		if d.receiver != "T" {
			t.Errorf("blank identifier ptr receiver: got receiver %q, want %q", d.receiver, "T")
		}
	} else {
		t.Errorf("expected key 'T.PtrMethod', got: %v", keys(decls))
	}
}

// TestAdversarial_UnicodeIdentifiers tests that Go source with Unicode
// identifiers is handled correctly by the AST parser and printer.
func TestAdversarial_UnicodeIdentifiers(t *testing.T) {
	src := []byte("package foo\n\nfunc \u03b1\u03b2\u03b3() int { return 42 }\n\ntype \u00c4nderung struct{ Wert int }\n")
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := decls["\u03b1\u03b2\u03b3"]; !ok {
		t.Errorf("expected Unicode function name key, got: %v", keys(decls))
	}
	if _, ok := decls["\u00c4nderung"]; !ok {
		t.Errorf("expected Unicode type name key, got: %v", keys(decls))
	}
}

// TestAdversarial_MixedVarBlock tests package-level var blocks with mixed types
// and values, including multi-name declarations.
func TestAdversarial_MixedVarBlock(t *testing.T) {
	src := []byte(`package foo

var (
	a int
	b string = "x"
	c, d float64
	e = 42
)
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "a", "b", "e" should each be their own entry.
	// "c, d" is a multi-name ValueSpec -- implementation takes Names[0] only.
	// This means "d" is silently lost. This is a potential defect.
	for _, expected := range []string{"a", "b", "e"} {
		if _, ok := decls[expected]; !ok {
			t.Errorf("expected key %q, got: %v", expected, keys(decls))
		}
	}
	// Document the behavior for multi-name var declarations:
	if _, ok := decls["c"]; !ok {
		t.Errorf("expected key 'c' (first name from multi-name var), got: %v", keys(decls))
	}
	if _, ok := decls["d"]; ok {
		t.Logf("INFO: multi-name var 'd' IS tracked separately (good)")
	} else {
		t.Logf("DEFECT: multi-name var 'd' is NOT tracked -- only first name 'c' captured from 'c, d float64'")
	}
}

// TestAdversarial_MixedConstBlockIota tests const blocks with iota, which
// have implicit values via iota counter.
func TestAdversarial_MixedConstBlockIota(t *testing.T) {
	src := []byte(`package foo

type Color int

const (
	Red Color = iota
	Green
	Blue
)
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All three consts should be extracted individually
	for _, name := range []string{"Red", "Green", "Blue"} {
		if d, ok := decls[name]; ok {
			if d.kind != DeclConst {
				t.Errorf("%s should be DeclConst, got %s", name, d.kind)
			}
		} else {
			t.Errorf("expected key %q, got: %v", name, keys(decls))
		}
	}
	// Plus the type declaration
	if _, ok := decls["Color"]; !ok {
		t.Errorf("expected key 'Color', got: %v", keys(decls))
	}
}

// TestAdversarial_EmptyFuncBody tests functions with empty bodies to ensure
// bodyLines is 0 and bodyHash is consistent.
func TestAdversarial_EmptyFuncBody(t *testing.T) {
	src := []byte("package foo\n\nfunc Empty() {}\n")
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := decls["Empty"]
	// An empty body {} on one line: end.Line - start.Line = 0
	if d.bodyLines != 0 {
		t.Errorf("empty func body should have bodyLines=0, got %d", d.bodyLines)
	}
}

// TestAdversarial_InterfaceMethodBodyLines tests that interface method
// declarations (which have no body) have bodyLines=0.
func TestAdversarial_InterfaceMethodBodyLines(t *testing.T) {
	src := []byte(`package foo

type Doer interface {
	Do(ctx int) error
	Undo() error
}
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Interface methods should NOT be extracted as top-level declarations.
	// Only the interface type itself is a top-level declaration.
	if _, ok := decls["Do"]; ok {
		t.Error("interface method 'Do' should NOT be extracted as a top-level declaration")
	}
	if _, ok := decls["Doer"]; !ok {
		t.Errorf("expected interface type 'Doer', got: %v", keys(decls))
	}
}

// TestAdversarial_GenericMethodReceiver tests methods on generic types where
// the receiver uses type parameters: func (s *Set[T]) Add(v T).
func TestAdversarial_GenericMethodReceiver(t *testing.T) {
	src := []byte(`package foo

type Set[T comparable] struct {
	m map[T]bool
}

func (s *Set[T]) Add(v T) {
	s.m[v] = true
}

func (s Set[T]) Has(v T) bool {
	return s.m[v]
}
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The receiver name should be "Set" (extracted from the generic type)
	if d, ok := decls["Set.Add"]; ok {
		if d.receiver != "Set" {
			t.Errorf("generic receiver: got %q, want %q", d.receiver, "Set")
		}
	} else {
		t.Errorf("expected key 'Set.Add', got: %v", keys(decls))
	}
	if d, ok := decls["Set.Has"]; ok {
		if d.receiver != "Set" {
			t.Errorf("generic receiver: got %q, want %q", d.receiver, "Set")
		}
	} else {
		t.Errorf("expected key 'Set.Has', got: %v", keys(decls))
	}
}

// TestAdversarial_NestedPointerGenericReceiver tests deeply nested receiver
// types like *Set[T] where extractReceiverName must recurse through StarExpr
// and IndexExpr.
func TestAdversarial_NestedPointerGenericReceiver(t *testing.T) {
	src := []byte(`package foo

type Cache[K comparable, V any] struct{}

func (c *Cache[K, V]) Get(k K) (V, bool) {
	var zero V
	return zero, false
}
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Multi-type-param generic: *Cache[K, V] uses IndexListExpr
	if d, ok := decls["Cache.Get"]; ok {
		if d.receiver != "Cache" {
			t.Errorf("multi-param generic receiver: got %q, want %q", d.receiver, "Cache")
		}
	} else {
		t.Errorf("expected key 'Cache.Get', got: %v", keys(decls))
	}
}

// TestAdversarial_LargeFile tests performance with a synthetically large file
// to verify the parser does not blow up on files with many declarations.
func TestAdversarial_LargeFile(t *testing.T) {
	var b strings.Builder
	b.WriteString("package foo\n\n")
	const numFuncs = 500
	for i := range numFuncs {
		b.WriteString(fmt.Sprintf("func Func%d() { _ = %d }\n\n", i, i))
	}
	src := []byte(b.String())
	decls, err := extractDeclarations(src, "large.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decls) != numFuncs {
		t.Errorf("expected %d declarations, got %d", numFuncs, len(decls))
	}
}

// TestAdversarial_LargeFileDiffPerformance tests that diffing two large files
// completes in reasonable time (no quadratic behavior).
func TestAdversarial_LargeFileDiffPerformance(t *testing.T) {
	var oldB, newB strings.Builder
	oldB.WriteString("package foo\n\n")
	newB.WriteString("package foo\n\n")
	const numFuncs = 500
	for i := range numFuncs {
		oldB.WriteString(fmt.Sprintf("func Func%d() { _ = %d }\n\n", i, i))
		switch {
		case i%3 == 0:
			// Modify every 3rd function
			newB.WriteString(fmt.Sprintf("func Func%d() { _ = %d + 1 }\n\n", i, i))
		case i%5 == 0:
			// Delete every 5th function (by not including it)
			continue
		default:
			newB.WriteString(fmt.Sprintf("func Func%d() { _ = %d }\n\n", i, i))
		}
	}
	// Add some new functions
	for i := numFuncs; i < numFuncs+50; i++ {
		newB.WriteString(fmt.Sprintf("func Func%d() { _ = %d }\n\n", i, i))
	}

	diff, err := ComputeFileDiff([]byte(oldB.String()), []byte(newB.String()), "large.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff == nil {
		t.Fatal("expected non-nil diff for large modified file")
	}
	if len(diff.Changes) == 0 {
		t.Error("expected some changes in large file diff")
	}
	t.Logf("INFO: large file diff produced %d changes", len(diff.Changes))
}

// TestAdversarial_TypeAlias tests type aliases (type Foo = Bar) which use
// a different AST representation than type definitions.
func TestAdversarial_TypeAlias(t *testing.T) {
	src := []byte(`package foo

type MyInt = int
type MyString = string
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := decls["MyInt"]; !ok {
		t.Errorf("expected key 'MyInt', got: %v", keys(decls))
	}
	if _, ok := decls["MyString"]; !ok {
		t.Errorf("expected key 'MyString', got: %v", keys(decls))
	}
}

// TestAdversarial_MethodOnPointerAndValue tests that methods on both pointer
// and value receivers for the same type with the same name are handled.
// This would be a compile error in Go, but the AST parser should still handle it.
func TestAdversarial_MethodSameNameDifferentTypes(t *testing.T) {
	src := []byte(`package foo

type A struct{}
type B struct{}

func (a A) String() string { return "A" }
func (b B) String() string { return "B" }
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Both should be present with distinct keys: "A.String" and "B.String"
	if _, ok := decls["A.String"]; !ok {
		t.Errorf("expected key 'A.String', got: %v", keys(decls))
	}
	if _, ok := decls["B.String"]; !ok {
		t.Errorf("expected key 'B.String', got: %v", keys(decls))
	}
}

// TestAdversarial_BuildConstraintFile tests a file with a build constraint
// comment. The parser should handle this gracefully.
func TestAdversarial_BuildConstraintFile(t *testing.T) {
	src := []byte(`//go:build linux

package foo

func PlatformFunc() string { return "linux" }
`)
	decls, err := extractDeclarations(src, "test_linux.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := decls["PlatformFunc"]; !ok {
		t.Errorf("expected key 'PlatformFunc', got: %v", keys(decls))
	}
}

// TestAdversarial_CgoImport tests that a file with a cgo import block
// is handled gracefully. The parser should either parse it successfully
// or return a clear error.
func TestAdversarial_CgoImport(t *testing.T) {
	src := []byte(`package foo

// #include <stdlib.h>
import "C"

func CgoFunc() {
	_ = C.free
}
`)
	// This should parse without error since go/parser handles cgo pseudo-package.
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		// If it fails, that is acceptable -- the file would go to SkippedFiles.
		t.Logf("INFO: cgo file parse error (acceptable, would be skipped): %v", err)
		return
	}
	if _, ok := decls["CgoFunc"]; !ok {
		t.Errorf("expected key 'CgoFunc', got: %v", keys(decls))
	}
}

// TestAdversarial_FormatDeclSig_EmptySig tests the formatDeclSig function
// when sig is empty. This exercises the fallback path.
func TestAdversarial_FormatDeclSig_EmptySig(t *testing.T) {
	result := formatDeclSig(DeclFunc, "")
	// When sig is empty, formatDeclSig falls through to "kind sig" which
	// produces "func " (kind + space + empty). This is a minor formatting issue.
	if result == "" {
		t.Error("formatDeclSig with empty sig should not return empty string")
	}
	t.Logf("INFO: formatDeclSig(DeclFunc, \"\") = %q", result)
}

// TestAdversarial_FormatModifiedZeroDelta tests formatting a MODIFIED change
// with BodyDelta == 0 (same line count but different body hash).
func TestAdversarial_FormatModifiedZeroDelta(t *testing.T) {
	diff := &SemanticDiff{
		FromHash: "abc",
		ToHash:   "def",
		Files: []FileDiff{
			{
				Path: "test.go",
				Changes: []DeclChange{
					{Kind: DeclFunc, Name: "Foo", Change: ChangeModified,
						NewSig: "func Foo()", BodyDelta: 0},
				},
			},
		},
	}
	result := FormatSemanticDiff(diff)
	// When BodyDelta is 0, no "[body: ...]" suffix should appear
	if strings.Contains(result, "[body:") {
		t.Errorf("zero body delta should not produce body annotation, got: %q", result)
	}
	if !strings.Contains(result, "MODIFIED") {
		t.Errorf("output should contain MODIFIED, got: %q", result)
	}
}

// TestAdversarial_NegativeBodyDelta tests formatting with a negative BodyDelta
// (function body got shorter).
func TestAdversarial_NegativeBodyDelta(t *testing.T) {
	diff := &SemanticDiff{
		FromHash: "abc",
		ToHash:   "def",
		Files: []FileDiff{
			{
				Path: "test.go",
				Changes: []DeclChange{
					{Kind: DeclFunc, Name: "Foo", Change: ChangeModified,
						NewSig: "func Foo()", BodyDelta: -3},
				},
			},
		},
	}
	result := FormatSemanticDiff(diff)
	// Should show negative delta without double sign
	if !strings.Contains(result, "[body: -3 lines]") {
		t.Errorf("expected '[body: -3 lines]', got: %q", result)
	}
}

// TestAdversarial_DeclKeySortStability tests that the sort order is fully
// deterministic when there are methods and functions with overlapping names.
func TestAdversarial_DeclKeySortStability(t *testing.T) {
	old := map[string]declaration{}
	new := map[string]declaration{
		"B.Foo": {kind: DeclFunc, name: "Foo", receiver: "B", signature: "func (b B) Foo()"},
		"A.Foo": {kind: DeclFunc, name: "Foo", receiver: "A", signature: "func (a A) Foo()"},
		"Foo":   {kind: DeclFunc, name: "Foo", signature: "func Foo()"},
		"A.Bar": {kind: DeclFunc, name: "Bar", receiver: "A", signature: "func (a A) Bar()"},
		"B.Bar": {kind: DeclFunc, name: "Bar", receiver: "B", signature: "func (b B) Bar()"},
	}
	changes := diffDeclarations(old, new)
	if len(changes) != 5 {
		t.Fatalf("expected 5 changes, got %d", len(changes))
	}

	// All are DeclFunc, so sort is by name key:
	// A.Bar, A.Foo, B.Bar, B.Foo, Foo
	expectedKeys := []string{"A.Bar", "A.Foo", "B.Bar", "B.Foo", "Foo"}
	for i, c := range changes {
		key := c.Name
		if c.Receiver != "" {
			key = c.Receiver + "." + c.Name
		}
		if key != expectedKeys[i] {
			t.Errorf("changes[%d] key = %q, want %q", i, key, expectedKeys[i])
		}
	}
}

// TestAdversarial_VarWithFuncValue tests package-level var declarations that
// hold function values, ensuring they are tracked as vars not funcs.
func TestAdversarial_VarWithFuncValue(t *testing.T) {
	src := []byte(`package foo

var handler = func(x int) int {
	return x * 2
}
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := decls["handler"]
	if !ok {
		t.Fatalf("expected key 'handler', got: %v", keys(decls))
	}
	if d.kind != DeclVar {
		t.Errorf("var holding func should be DeclVar, got %s", d.kind)
	}
}

// TestAdversarial_ConstGroupImplicitValue tests const groups where some
// entries have implicit values (iota continuation).
func TestAdversarial_ConstGroupImplicitValue(t *testing.T) {
	src := []byte(`package foo

const (
	A = iota
	B
	C
	D = 100
	E = iota
)
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"A", "B", "C", "D", "E"}
	for _, name := range expected {
		if d, ok := decls[name]; !ok {
			t.Errorf("missing expected const %q, got: %v", name, keys(decls))
		} else if d.kind != DeclConst {
			t.Errorf("%s should be DeclConst, got %s", name, d.kind)
		}
	}
}

// TestAdversarial_PackageLevelBlankVar tests that package-level blank
// identifier vars (var _ Interface = (*Impl)(nil)) are extracted.
// This is a common Go pattern for compile-time interface checks.
func TestAdversarial_PackageLevelBlankVar(t *testing.T) {
	src := []byte(`package foo

type Stringer interface{ String() string }
type MyType struct{}
func (m MyType) String() string { return "" }

var _ Stringer = (*MyType)(nil)
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "_" is a valid key for blank identifier vars.
	// Multiple blank vars would collide on key "_" (same issue as init()).
	if _, ok := decls["_"]; !ok {
		t.Logf("INFO: blank identifier var not tracked (key '_' missing)")
	} else {
		t.Logf("INFO: blank identifier var IS tracked with key '_'")
	}
	// No panic is the primary assertion
}

// TestAdversarial_EmbeddedStruct tests type declarations with embedded fields.
func TestAdversarial_EmbeddedStruct(t *testing.T) {
	src := []byte(`package foo

type Base struct{ Name string }

type Extended struct {
	Base
	Extra int
}
`)
	decls, err := extractDeclarations(src, "test.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := decls["Base"]; !ok {
		t.Errorf("expected key 'Base', got: %v", keys(decls))
	}
	if _, ok := decls["Extended"]; !ok {
		t.Errorf("expected key 'Extended', got: %v", keys(decls))
	}
}
