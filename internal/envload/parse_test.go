package envload

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "simple key-value",
			input: "FOO=bar",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "multiple key-values",
			input: "FOO=bar\nBAZ=qux",
			want:  map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name:  "comments ignored",
			input: "# this is a comment\nFOO=bar",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "empty lines ignored",
			input: "\nFOO=bar\n\nBAZ=qux\n",
			want:  map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name:  "double-quoted value",
			input: `FOO="hello world"`,
			want:  map[string]string{"FOO": "hello world"},
		},
		{
			name:  "double-quoted with newline escape",
			input: `FOO="line1\nline2"`,
			want:  map[string]string{"FOO": "line1\nline2"},
		},
		{
			name:  "double-quoted with tab escape",
			input: `FOO="col1\tcol2"`,
			want:  map[string]string{"FOO": "col1\tcol2"},
		},
		{
			name:  "double-quoted with escaped backslash",
			input: `FOO="path\\to\\file"`,
			want:  map[string]string{"FOO": `path\to\file`},
		},
		{
			name:  "double-quoted with escaped quote",
			input: `FOO="say \"hello\""`,
			want:  map[string]string{"FOO": `say "hello"`},
		},
		{
			name:  "single-quoted value literal",
			input: `FOO='hello\nworld'`,
			want:  map[string]string{"FOO": `hello\nworld`},
		},
		{
			name:  "empty value",
			input: "FOO=",
			want:  map[string]string{"FOO": ""},
		},
		{
			name:  "value with equals sign",
			input: "FOO=bar=baz",
			want:  map[string]string{"FOO": "bar=baz"},
		},
		{
			name:  "value with hash (no trailing comments)",
			input: "FOO=abc#123",
			want:  map[string]string{"FOO": "abc#123"},
		},
		{
			name:  "leading whitespace on key trimmed",
			input: "  FOO=bar",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "trailing whitespace on unquoted value trimmed",
			input: "FOO=bar  ",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "whitespace around equals",
			input: "FOO = bar",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "empty input",
			input: "",
			want:  map[string]string{},
		},
		{
			name:  "only comments",
			input: "# comment 1\n# comment 2",
			want:  map[string]string{},
		},
		{
			name:  "duplicate key uses last value",
			input: "FOO=first\nFOO=second",
			want:  map[string]string{"FOO": "second"},
		},
		{
			name:    "malformed line no equals",
			input:   "BADLINE",
			wantErr: true,
		},
		{
			name:    "unterminated double quote",
			input:   `FOO="unterminated`,
			wantErr: true,
		},
		{
			name:    "unterminated single quote",
			input:   `FOO='unterminated`,
			wantErr: true,
		},
		{
			name:  "dollar sign treated as literal",
			input: "FOO=$BAR",
			want:  map[string]string{"FOO": "$BAR"},
		},
		{
			name:  "real-world secrets",
			input: "SLACK_SIGNING_SECRET=abc123def456\nSLACK_BOT_TOKEN=xoxb-123-456-abc\nANTHROPIC_API_KEY=sk-ant-xxx",
			want: map[string]string{
				"SLACK_SIGNING_SECRET": "abc123def456",
				"SLACK_BOT_TOKEN":     "xoxb-123-456-abc",
				"ANTHROPIC_API_KEY":   "sk-ant-xxx",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Parse() got %d keys, want %d", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("Parse() key %q = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}
