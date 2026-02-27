package mena

import "testing"

func TestStripMenaExtension(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"INDEX.dro.md", "INDEX.md"},
		{"INDEX.lego.md", "INDEX.md"},
		{"commit.dro.md", "commit.md"},
		{"prompting.lego.md", "prompting.md"},
		{"helper.md", "helper.md"},
		{"README.md", "README.md"},
		{"data.json", "data.json"},
		{"foo.dro.dro.md", "foo.dro.md"}, // only first infix stripped
	}
	for _, tc := range cases {
		got := StripMenaExtension(tc.input)
		if got != tc.expected {
			t.Errorf("StripMenaExtension(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestDetectMenaType(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"INDEX.dro.md", "dro"},
		{"INDEX.lego.md", "lego"},
		{"commit.dro.md", "dro"},
		{"prompting.lego.md", "lego"},
		{"helper.md", "dro"}, // default
		{"README.md", "dro"}, // default
	}
	for _, tc := range cases {
		got := DetectMenaType(tc.input)
		if got != tc.expected {
			t.Errorf("DetectMenaType(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestRouteMenaFile(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"INDEX.dro.md", "commands"},
		{"INDEX.lego.md", "skills"},
		{"commit.dro.md", "commands"},
		{"prompting.lego.md", "skills"},
		{"helper.md", "commands"}, // default is dro -> commands
	}
	for _, tc := range cases {
		got := RouteMenaFile(tc.input)
		if got != tc.expected {
			t.Errorf("RouteMenaFile(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
