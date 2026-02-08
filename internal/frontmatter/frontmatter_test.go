package frontmatter

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParse_ValidFrontmatter(t *testing.T) {
	content := []byte("---\nname: test\ndescription: hello\n---\n\n# Body\n")

	yamlBytes, body, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(yamlBytes) != "name: test\ndescription: hello" {
		t.Errorf("yamlBytes = %q, want %q", string(yamlBytes), "name: test\ndescription: hello")
	}
	if string(body) != "\n# Body\n" {
		t.Errorf("body = %q, want %q", string(body), "\n# Body\n")
	}
}

func TestParse_CRLFDelimiters(t *testing.T) {
	content := []byte("---\r\nname: test\r\n---\r\n\r\n# Body\r\n")

	yamlBytes, _, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// YAML bytes preserve raw content including \r from CRLF line endings.
	// The YAML parser handles this transparently during Unmarshal.
	if string(yamlBytes) != "name: test\r" {
		t.Errorf("yamlBytes = %q, want %q", string(yamlBytes), "name: test\r")
	}
}

func TestParse_MixedLineEndings(t *testing.T) {
	// Open with \r\n, close with \n --- YAML bytes preserve raw content
	content := []byte("---\r\nname: test\n---\n\n# Body\n")

	yamlBytes, _, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(yamlBytes) != "name: test" {
		t.Errorf("yamlBytes = %q, want %q", string(yamlBytes), "name: test")
	}
}

func TestParse_MissingOpenDelimiter(t *testing.T) {
	content := []byte("name: test\n---\n")

	_, _, err := Parse(content)
	if err != ErrMissingOpenDelimiter {
		t.Errorf("err = %v, want ErrMissingOpenDelimiter", err)
	}
}

func TestParse_MissingCloseDelimiter(t *testing.T) {
	content := []byte("---\nname: test\nno closing\n")

	_, _, err := Parse(content)
	if err != ErrMissingCloseDelimiter {
		t.Errorf("err = %v, want ErrMissingCloseDelimiter", err)
	}
}

func TestParse_EmptyContent(t *testing.T) {
	_, _, err := Parse([]byte(""))
	if err != ErrMissingOpenDelimiter {
		t.Errorf("err = %v, want ErrMissingOpenDelimiter", err)
	}
}

func TestParse_EmptyFrontmatter(t *testing.T) {
	content := []byte("---\n\n---\n\n# Body\n")

	yamlBytes, body, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(yamlBytes) != "" {
		t.Errorf("yamlBytes = %q, want empty", string(yamlBytes))
	}
	if string(body) != "\n# Body\n" {
		t.Errorf("body = %q, want %q", string(body), "\n# Body\n")
	}
}

func TestParse_NoBody(t *testing.T) {
	content := []byte("---\nname: test\n---\n")

	yamlBytes, body, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(yamlBytes) != "name: test" {
		t.Errorf("yamlBytes = %q, want %q", string(yamlBytes), "name: test")
	}
	if string(body) != "" {
		t.Errorf("body = %q, want empty", string(body))
	}
}

func TestFlexibleStringSlice_CommaSeparated(t *testing.T) {
	input := "Bash, Read, Glob, Grep"

	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: input,
	}

	var fs FlexibleStringSlice
	if err := fs.UnmarshalYAML(node); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"Bash", "Read", "Glob", "Grep"}
	if len(fs) != len(expected) {
		t.Fatalf("len = %d, want %d", len(fs), len(expected))
	}
	for i, v := range fs {
		if v != expected[i] {
			t.Errorf("fs[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestFlexibleStringSlice_YAMLSequence(t *testing.T) {
	yamlInput := "- Bash\n- Read\n- Glob\n"
	var fs FlexibleStringSlice
	if err := yaml.Unmarshal([]byte(yamlInput), &fs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"Bash", "Read", "Glob"}
	if len(fs) != len(expected) {
		t.Fatalf("len = %d, want %d", len(fs), len(expected))
	}
	for i, v := range fs {
		if v != expected[i] {
			t.Errorf("fs[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestFlexibleStringSlice_EmptyString(t *testing.T) {
	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "",
	}

	var fs FlexibleStringSlice
	if err := fs.UnmarshalYAML(node); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fs) != 0 {
		t.Errorf("len = %d, want 0; got %v", len(fs), fs)
	}
}

func TestFlexibleStringSlice_SingleValue(t *testing.T) {
	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "Read",
	}

	var fs FlexibleStringSlice
	if err := fs.UnmarshalYAML(node); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fs) != 1 || fs[0] != "Read" {
		t.Errorf("fs = %v, want [Read]", fs)
	}
}

func TestFlexibleStringSlice_TrailingComma(t *testing.T) {
	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "Bash, Read,",
	}

	var fs FlexibleStringSlice
	if err := fs.UnmarshalYAML(node); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Trailing comma produces empty part, which gets filtered out
	expected := []string{"Bash", "Read"}
	if len(fs) != len(expected) {
		t.Fatalf("len = %d, want %d; got %v", len(fs), len(expected), fs)
	}
	for i, v := range fs {
		if v != expected[i] {
			t.Errorf("fs[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestFlexibleStringSlice_RoundTrip(t *testing.T) {
	// Parse frontmatter containing FlexibleStringSlice field, verify it works
	// end-to-end with Parse() + yaml.Unmarshal
	content := []byte("---\ntools: Bash, Read, Glob\nname: test\n---\n\n# Body\n")

	yamlBytes, _, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	var result struct {
		Tools FlexibleStringSlice `yaml:"tools"`
		Name  string              `yaml:"name"`
	}
	if err := yaml.Unmarshal(yamlBytes, &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if result.Name != "test" {
		t.Errorf("name = %q, want %q", result.Name, "test")
	}
	expected := []string{"Bash", "Read", "Glob"}
	if len(result.Tools) != len(expected) {
		t.Fatalf("tools len = %d, want %d", len(result.Tools), len(expected))
	}
	for i, v := range result.Tools {
		if v != expected[i] {
			t.Errorf("tools[%d] = %q, want %q", i, v, expected[i])
		}
	}
}
