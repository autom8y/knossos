package inscription

import (
	"strings"
	"testing"
)

func TestMarkerParser_Parse_ValidMarkers(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		wantRegions    int
		wantMarkers    int
		wantErrors     int
		checkRegion    string
		checkDirective MarkerDirective
	}{
		{
			name: "simple START/END region",
			content: `# CLAUDE.md
<!-- KNOSSOS:START quick-start -->
## Quick Start
This is the content.
<!-- KNOSSOS:END quick-start -->
`,
			wantRegions:    1,
			wantMarkers:    2,
			wantErrors:     0,
			checkRegion:    "quick-start",
			checkDirective: DirectiveStart,
		},
		{
			name: "multiple regions",
			content: `<!-- KNOSSOS:START section-a -->
Content A
<!-- KNOSSOS:END section-a -->

<!-- KNOSSOS:START section-b -->
Content B
<!-- KNOSSOS:END section-b -->
`,
			wantRegions: 2,
			wantMarkers: 4,
			wantErrors:  0,
		},
		{
			name: "anchor directive",
			content: `# Header
<!-- KNOSSOS:ANCHOR agent-table -->
Some content after
`,
			wantRegions:    1,
			wantMarkers:    1,
			wantErrors:     0,
			checkRegion:    "agent-table",
			checkDirective: DirectiveAnchor,
		},
		{
			name: "marker with options",
			content: `<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE -->
## Quick Start
<!-- KNOSSOS:END quick-start -->
`,
			wantRegions: 1,
			wantMarkers: 2,
			wantErrors:  0,
		},
		{
			name: "region name with numbers",
			content: `<!-- KNOSSOS:START section-2-alpha -->
Content
<!-- KNOSSOS:END section-2-alpha -->
`,
			wantRegions:    1,
			wantMarkers:    2,
			wantErrors:     0,
			checkRegion:    "section-2-alpha",
			checkDirective: DirectiveStart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewMarkerParser()
			result := p.Parse(tt.content)

			if len(result.Regions) != tt.wantRegions {
				t.Errorf("Parse() got %d regions, want %d", len(result.Regions), tt.wantRegions)
			}

			if len(result.Markers) != tt.wantMarkers {
				t.Errorf("Parse() got %d markers, want %d", len(result.Markers), tt.wantMarkers)
			}

			if len(result.Errors) != tt.wantErrors {
				t.Errorf("Parse() got %d errors, want %d: %v", len(result.Errors), tt.wantErrors, result.Errors)
			}

			if tt.checkRegion != "" {
				region := result.GetRegion(tt.checkRegion)
				if region == nil {
					t.Errorf("Parse() region %q not found", tt.checkRegion)
				} else if region.StartMarker.Directive != tt.checkDirective {
					t.Errorf("Parse() region %q directive = %v, want %v",
						tt.checkRegion, region.StartMarker.Directive, tt.checkDirective)
				}
			}
		})
	}
}

func TestMarkerParser_Parse_Options(t *testing.T) {
	content := `<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE owner=regenerate -->
Content
<!-- KNOSSOS:END quick-start -->
`
	p := NewMarkerParser()
	result := p.Parse(content)

	if len(result.Errors) > 0 {
		t.Fatalf("Parse() unexpected errors: %v", result.Errors)
	}

	region := result.GetRegion("quick-start")
	if region == nil {
		t.Fatal("Parse() region 'quick-start' not found")
	}

	marker := region.StartMarker
	tests := []struct {
		key   string
		value string
	}{
		{"regenerate", "true"},
		{"source", "ACTIVE_RITE"},
		{"owner", "regenerate"},
	}

	for _, tt := range tests {
		if got := marker.GetOption(tt.key); got != tt.value {
			t.Errorf("GetOption(%q) = %q, want %q", tt.key, got, tt.value)
		}
	}
}

func TestMarkerParser_Parse_EscapeMechanisms(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantRegions int
		wantMarkers int
	}{
		{
			name:        "markers in code block are ignored",
			content:     "# Header\n```\n<!-- KNOSSOS:START test -->\nContent\n<!-- KNOSSOS:END test -->\n```\n",
			wantRegions: 0,
			wantMarkers: 0,
		},
		{
			name: "backslash escaped marker",
			content: `# Header
\<!-- KNOSSOS:START test -->
Content
`,
			wantRegions: 0,
			wantMarkers: 0,
		},
		{
			name: "HTML entity escaped marker",
			content: `# Header
&lt;!-- KNOSSOS:START test --&gt;
Content
`,
			wantRegions: 0,
			wantMarkers: 0,
		},
		{
			name:        "real marker outside code block",
			content:     "```\nCode block content\n```\n<!-- KNOSSOS:START test -->\nContent\n<!-- KNOSSOS:END test -->\n",
			wantRegions: 1,
			wantMarkers: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewMarkerParser()
			result := p.Parse(tt.content)

			if len(result.Regions) != tt.wantRegions {
				t.Errorf("Parse() got %d regions, want %d", len(result.Regions), tt.wantRegions)
			}

			if len(result.Markers) != tt.wantMarkers {
				t.Errorf("Parse() got %d markers, want %d", len(result.Markers), tt.wantMarkers)
			}
		})
	}
}

func TestMarkerParser_Parse_Errors(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErrors  int
		errorSubstr string
	}{
		{
			name: "START without END",
			content: `<!-- KNOSSOS:START test -->
Content without closing tag
`,
			wantErrors:  1,
			errorSubstr: "without matching END",
		},
		{
			name: "END without START",
			content: `Content
<!-- KNOSSOS:END test -->
`,
			wantErrors:  1,
			errorSubstr: "without matching START",
		},
		{
			name: "nested regions (not allowed)",
			content: `<!-- KNOSSOS:START outer -->
Outer content
<!-- KNOSSOS:START inner -->
Inner content
<!-- KNOSSOS:END inner -->
<!-- KNOSSOS:END outer -->
`,
			wantErrors:  2, // One for nested START, one for orphan END
			errorSubstr: "nested START",
		},
		{
			name: "duplicate region name",
			content: `<!-- KNOSSOS:START test -->
First
<!-- KNOSSOS:END test -->
<!-- KNOSSOS:START test -->
Second
<!-- KNOSSOS:END test -->
`,
			wantErrors:  1,
			errorSubstr: "duplicate region",
		},
		{
			name: "duplicate anchor",
			content: `<!-- KNOSSOS:ANCHOR test -->
<!-- KNOSSOS:ANCHOR test -->
`,
			wantErrors:  1,
			errorSubstr: "duplicate region",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewMarkerParser()
			result := p.Parse(tt.content)

			if len(result.Errors) != tt.wantErrors {
				t.Errorf("Parse() got %d errors, want %d: %v",
					len(result.Errors), tt.wantErrors, result.Errors)
				return
			}

			if tt.errorSubstr != "" && len(result.Errors) > 0 {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err.Message, tt.errorSubstr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Parse() error should contain %q, got %v", tt.errorSubstr, result.Errors)
				}
			}
		})
	}
}

func TestMarkerParser_Parse_RegionContent(t *testing.T) {
	content := `<!-- KNOSSOS:START test -->
Line 1
Line 2
Line 3
<!-- KNOSSOS:END test -->
`
	p := NewMarkerParser()
	result := p.Parse(content)

	region := result.GetRegion("test")
	if region == nil {
		t.Fatal("Parse() region 'test' not found")
	}

	expectedContent := "Line 1\nLine 2\nLine 3"
	if region.Content != expectedContent {
		t.Errorf("Parse() content = %q, want %q", region.Content, expectedContent)
	}

	if region.StartLine != 1 {
		t.Errorf("Parse() StartLine = %d, want 1", region.StartLine)
	}

	if region.EndLine != 5 {
		t.Errorf("Parse() EndLine = %d, want 5", region.EndLine)
	}

	if region.LineCount() != 3 {
		t.Errorf("Parse() LineCount() = %d, want 3", region.LineCount())
	}
}

func TestMarkerParser_ParseLegacyMarkers(t *testing.T) {
	content := `# Header
<!-- PRESERVE: satellite-owned -->
Some content
<!-- SYNC: roster-owned -->
More content
`
	p := NewMarkerParser()
	legacy := p.ParseLegacyMarkers(content)

	if len(legacy) != 2 {
		t.Fatalf("ParseLegacyMarkers() got %d markers, want 2", len(legacy))
	}

	if legacy[0].Type != LegacyPreserve {
		t.Errorf("ParseLegacyMarkers() [0].Type = %v, want %v", legacy[0].Type, LegacyPreserve)
	}

	if legacy[1].Type != LegacySync {
		t.Errorf("ParseLegacyMarkers() [1].Type = %v, want %v", legacy[1].Type, LegacySync)
	}

	// Check suggested region names
	if legacy[0].SuggestedRegionName == "" {
		t.Error("ParseLegacyMarkers() [0].SuggestedRegionName should not be empty")
	}
}

func TestMarkerParser_ParseLegacyMarkers_InCodeBlock(t *testing.T) {
	content := "# Header\n```\n<!-- PRESERVE: satellite-owned -->\n```\n"
	p := NewMarkerParser()
	legacy := p.ParseLegacyMarkers(content)

	if len(legacy) != 0 {
		t.Errorf("ParseLegacyMarkers() should ignore markers in code blocks, got %d", len(legacy))
	}
}

func TestValidateRegionName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "test", false},
		{"valid with hyphen", "quick-start", false},
		{"valid with number", "section-2", false},
		{"valid complex", "agent-routing-config", false},

		{"empty", "", true},
		{"starts with number", "2test", true},
		{"starts with hyphen", "-test", true},
		{"ends with hyphen", "test-", true},
		{"consecutive hyphens", "test--name", true},
		{"uppercase", "Test", true},
		{"underscore", "test_name", true},
		{"space", "test name", true},
		{"special char", "test@name", true},
		{"too long", strings.Repeat("a", 65), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegionName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRegionName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestBuildMarker(t *testing.T) {
	tests := []struct {
		name       string
		directive  MarkerDirective
		regionName string
		options    map[string]string
		want       string
	}{
		{
			name:       "START without options",
			directive:  DirectiveStart,
			regionName: "test",
			options:    nil,
			want:       "<!-- KNOSSOS:START test -->",
		},
		{
			name:       "END marker",
			directive:  DirectiveEnd,
			regionName: "test",
			options:    nil,
			want:       "<!-- KNOSSOS:END test -->",
		},
		{
			name:       "ANCHOR marker",
			directive:  DirectiveAnchor,
			regionName: "agent-table",
			options:    nil,
			want:       "<!-- KNOSSOS:ANCHOR agent-table -->",
		},
		{
			name:       "START with options",
			directive:  DirectiveStart,
			regionName: "quick-start",
			options:    map[string]string{"regenerate": "true"},
			want:       "<!-- KNOSSOS:START quick-start regenerate=true -->",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildMarker(tt.directive, tt.regionName, tt.options)
			if got != tt.want {
				t.Errorf("BuildMarker() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildStartEndMarkers(t *testing.T) {
	start := BuildStartMarker("test", nil)
	if start != "<!-- KNOSSOS:START test -->" {
		t.Errorf("BuildStartMarker() = %q", start)
	}

	end := BuildEndMarker("test")
	if end != "<!-- KNOSSOS:END test -->" {
		t.Errorf("BuildEndMarker() = %q", end)
	}

	anchor := BuildAnchorMarker("test")
	if anchor != "<!-- KNOSSOS:ANCHOR test -->" {
		t.Errorf("BuildAnchorMarker() = %q", anchor)
	}
}

func TestWrapContent(t *testing.T) {
	content := "Line 1\nLine 2"
	wrapped := WrapContent("test", content, nil)

	expected := "<!-- KNOSSOS:START test -->\nLine 1\nLine 2\n<!-- KNOSSOS:END test -->"
	if wrapped != expected {
		t.Errorf("WrapContent() = %q, want %q", wrapped, expected)
	}
}

func TestWrapContent_WithTrailingNewline(t *testing.T) {
	content := "Line 1\nLine 2\n"
	wrapped := WrapContent("test", content, nil)

	// Should not add extra newline
	expected := "<!-- KNOSSOS:START test -->\nLine 1\nLine 2\n<!-- KNOSSOS:END test -->"
	if wrapped != expected {
		t.Errorf("WrapContent() = %q, want %q", wrapped, expected)
	}
}

func TestMarkerDirective_IsValid(t *testing.T) {
	tests := []struct {
		directive MarkerDirective
		want      bool
	}{
		{DirectiveStart, true},
		{DirectiveEnd, true},
		{DirectiveAnchor, true},
		{MarkerDirective("INVALID"), false},
		{MarkerDirective(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.directive), func(t *testing.T) {
			if got := tt.directive.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarkerDirective_RequiresEnd(t *testing.T) {
	tests := []struct {
		directive MarkerDirective
		want      bool
	}{
		{DirectiveStart, true},
		{DirectiveEnd, false},
		{DirectiveAnchor, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.directive), func(t *testing.T) {
			if got := tt.directive.RequiresEnd(); got != tt.want {
				t.Errorf("RequiresEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOwnerType_IsValid(t *testing.T) {
	tests := []struct {
		owner OwnerType
		want  bool
	}{
		{OwnerKnossos, true},
		{OwnerSatellite, true},
		{OwnerRegenerate, true},
		{OwnerType("invalid"), false},
		{OwnerType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.owner), func(t *testing.T) {
			if got := tt.owner.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarker_GetOption_HasOption(t *testing.T) {
	marker := &Marker{
		Options: map[string]string{
			"regenerate": "true",
			"source":     "ACTIVE_RITE",
		},
	}

	if got := marker.GetOption("regenerate"); got != "true" {
		t.Errorf("GetOption('regenerate') = %q, want 'true'", got)
	}

	if got := marker.GetOption("missing"); got != "" {
		t.Errorf("GetOption('missing') = %q, want ''", got)
	}

	if !marker.HasOption("regenerate") {
		t.Error("HasOption('regenerate') = false, want true")
	}

	if marker.HasOption("missing") {
		t.Error("HasOption('missing') = true, want false")
	}
}

func TestMarker_GetOption_NilOptions(t *testing.T) {
	marker := &Marker{
		Options: nil,
	}

	if got := marker.GetOption("any"); got != "" {
		t.Errorf("GetOption('any') on nil options = %q, want ''", got)
	}

	if marker.HasOption("any") {
		t.Error("HasOption('any') on nil options = true, want false")
	}
}

func TestParsedRegion_IsAnchor(t *testing.T) {
	anchorRegion := &ParsedRegion{
		StartMarker: &Marker{Directive: DirectiveAnchor},
	}
	if !anchorRegion.IsAnchor() {
		t.Error("IsAnchor() = false for anchor region, want true")
	}

	startEndRegion := &ParsedRegion{
		StartMarker: &Marker{Directive: DirectiveStart},
		EndMarker:   &Marker{Directive: DirectiveEnd},
	}
	if startEndRegion.IsAnchor() {
		t.Error("IsAnchor() = true for START/END region, want false")
	}
}

func TestParseResult_HasErrors(t *testing.T) {
	resultNoErrors := &ParseResult{
		Errors: []ParseError{},
	}
	if resultNoErrors.HasErrors() {
		t.Error("HasErrors() = true with no errors, want false")
	}

	resultWithErrors := &ParseResult{
		Errors: []ParseError{{Message: "test error"}},
	}
	if !resultWithErrors.HasErrors() {
		t.Error("HasErrors() = false with errors, want true")
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{100, "100"},
		{-1, "-1"},
		{-42, "-42"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := itoa(tt.input); got != tt.want {
				t.Errorf("itoa(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseError_Error(t *testing.T) {
	pe := ParseError{
		Line:    10,
		Message: "test error message",
	}

	if got := pe.Error(); got != "test error message" {
		t.Errorf("Error() = %q, want %q", got, "test error message")
	}
}

func TestParsedRegion_LineCount(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
	}{
		{"empty", "", 0},
		{"single line", "hello", 1},
		{"two lines", "hello\nworld", 2},
		{"three lines", "a\nb\nc", 3},
		{"trailing newline", "hello\n", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &ParsedRegion{Content: tt.content}
			if got := pr.LineCount(); got != tt.want {
				t.Errorf("LineCount() = %d, want %d", got, tt.want)
			}
		})
	}
}
