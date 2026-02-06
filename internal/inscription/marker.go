package inscription

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
)

// Marker syntax constants
const (
	// MarkerPrefix is the namespace prefix for KNOSSOS markers.
	MarkerPrefix = "KNOSSOS:"

	// MarkerOpenTag is the HTML comment opening.
	MarkerOpenTag = "<!--"

	// MarkerCloseTag is the HTML comment closing.
	MarkerCloseTag = "-->"
)

// markerRegex matches KNOSSOS markers in the format:
// <!-- KNOSSOS:{DIRECTIVE} {REGION_NAME} [{OPTIONS}] -->
// Groups: 1=directive, 2=region_name, 3=options (optional)
var markerRegex = regexp.MustCompile(
	`<!--\s*KNOSSOS:(START|END|ANCHOR)\s+([a-z][a-z0-9-]*(?:-[a-z0-9]+)*)(?:\s+([^>]+?))?\s*-->`,
)

// optionRegex matches key=value pairs in marker options.
var optionRegex = regexp.MustCompile(`([a-z_][a-z0-9_]*)=([^\s]+)`)

// legacyPreserveRegex matches legacy PRESERVE markers.
var legacyPreserveRegex = regexp.MustCompile(`<!--\s*PRESERVE:\s*([^>]+?)\s*-->`)

// legacySyncRegex matches legacy SYNC markers.
var legacySyncRegex = regexp.MustCompile(`<!--\s*SYNC:\s*([^>]+?)\s*-->`)

// codeBlockStartRegex matches the start of a fenced code block.
var codeBlockStartRegex = regexp.MustCompile("^```")

// codeBlockEndRegex matches the end of a fenced code block.
var codeBlockEndRegex = regexp.MustCompile("^```\\s*$")

// MarkerParser parses KNOSSOS markers from CLAUDE.md content.
type MarkerParser struct {
	// IgnoreLegacy skips legacy marker detection if true.
	IgnoreLegacy bool

	// StrictMode enables stricter validation if true.
	StrictMode bool
}

// NewMarkerParser creates a new marker parser with default settings.
func NewMarkerParser() *MarkerParser {
	return &MarkerParser{}
}

// Parse parses all KNOSSOS markers from the given content.
// Returns a ParseResult containing regions, markers, and any errors.
func (p *MarkerParser) Parse(content string) *ParseResult {
	result := &ParseResult{
		Regions: make(map[string]*ParsedRegion),
		Markers: make([]*Marker, 0),
		Errors:  make([]ParseError, 0),
	}

	lines := strings.Split(content, "\n")
	inCodeBlock := false
	openRegions := make(map[string]*Marker) // Track START markers awaiting END

	for lineNum, line := range lines {
		lineNumber := lineNum + 1 // 1-indexed

		// Track code block state for escape mechanism
		if codeBlockStartRegex.MatchString(strings.TrimSpace(line)) {
			if inCodeBlock {
				// Check if this is actually an end marker
				if codeBlockEndRegex.MatchString(strings.TrimSpace(line)) {
					inCodeBlock = false
				}
			} else {
				inCodeBlock = true
			}
			continue
		}

		// Skip markers inside code blocks (escape mechanism)
		if inCodeBlock {
			continue
		}

		// Check for escaped markers (backslash escape)
		if strings.Contains(line, "\\<!-- KNOSSOS:") {
			continue
		}

		// Check for HTML entity escaped markers
		if strings.Contains(line, "&lt;!-- KNOSSOS:") {
			continue
		}

		// Try to parse KNOSSOS marker
		marker := p.parseMarkerLine(line, lineNumber)
		if marker != nil {
			result.Markers = append(result.Markers, marker)

			// Process marker based on directive
			switch marker.Directive {
			case DirectiveStart:
				// Check for nested region (error)
				if existing, ok := openRegions[marker.RegionName]; ok {
					result.Errors = append(result.Errors, ParseError{
						Line:    lineNumber,
						Message: "nested START marker for region '" + marker.RegionName + "' (already started at line " + itoa(existing.LineNumber) + ")",
						Raw:     marker.Raw,
					})
					continue
				}

				// Check for any open region (no nesting allowed)
				for regionName, existing := range openRegions {
					result.Errors = append(result.Errors, ParseError{
						Line:    lineNumber,
						Message: "nested START marker '" + marker.RegionName + "' inside open region '" + regionName + "' (started at line " + itoa(existing.LineNumber) + ")",
						Raw:     marker.Raw,
					})
				}
				if len(openRegions) > 0 {
					continue
				}

				openRegions[marker.RegionName] = marker

			case DirectiveEnd:
				// Check for matching START
				startMarker, ok := openRegions[marker.RegionName]
				if !ok {
					result.Errors = append(result.Errors, ParseError{
						Line:    lineNumber,
						Message: "END marker for region '" + marker.RegionName + "' without matching START",
						Raw:     marker.Raw,
					})
					continue
				}

				// Extract content between markers
				contentLines := lines[startMarker.LineNumber : lineNumber-1]
				content := strings.Join(contentLines, "\n")

				// Create parsed region
				region := &ParsedRegion{
					Name:        marker.RegionName,
					StartMarker: startMarker,
					EndMarker:   marker,
					Content:     content,
					StartLine:   startMarker.LineNumber,
					EndLine:     lineNumber,
				}

				// Check for duplicate region
				if _, exists := result.Regions[marker.RegionName]; exists {
					result.Errors = append(result.Errors, ParseError{
						Line:    lineNumber,
						Message: "duplicate region '" + marker.RegionName + "'",
						Raw:     marker.Raw,
					})
				} else {
					result.Regions[marker.RegionName] = region
				}

				delete(openRegions, marker.RegionName)

			case DirectiveAnchor:
				// Anchor is a single-line insertion point
				region := &ParsedRegion{
					Name:        marker.RegionName,
					StartMarker: marker,
					EndMarker:   nil,
					Content:     "",
					StartLine:   lineNumber,
					EndLine:     0,
				}

				// Check for duplicate region
				if _, exists := result.Regions[marker.RegionName]; exists {
					result.Errors = append(result.Errors, ParseError{
						Line:    lineNumber,
						Message: "duplicate region '" + marker.RegionName + "'",
						Raw:     marker.Raw,
					})
				} else {
					result.Regions[marker.RegionName] = region
				}
			}
		}
	}

	// Check for unclosed regions
	for regionName, startMarker := range openRegions {
		result.Errors = append(result.Errors, ParseError{
			Line:    startMarker.LineNumber,
			Message: "START marker for region '" + regionName + "' without matching END",
			Raw:     startMarker.Raw,
		})
	}

	return result
}

// parseMarkerLine attempts to parse a KNOSSOS marker from a single line.
// Returns nil if the line does not contain a valid marker.
func (p *MarkerParser) parseMarkerLine(line string, lineNumber int) *Marker {
	matches := markerRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	directive := MarkerDirective(matches[1])
	regionName := matches[2]
	optionsStr := ""
	if len(matches) > 3 {
		optionsStr = strings.TrimSpace(matches[3])
	}

	// Parse options
	options := make(map[string]string)
	if optionsStr != "" {
		optMatches := optionRegex.FindAllStringSubmatch(optionsStr, -1)
		for _, match := range optMatches {
			if len(match) >= 3 {
				key := match[1]
				value := strings.Trim(match[2], `"'`)
				options[key] = value
			}
		}
	}

	return &Marker{
		Directive:  directive,
		RegionName: regionName,
		Options:    options,
		LineNumber: lineNumber,
		Raw:        strings.TrimSpace(matches[0]),
	}
}

// ParseLegacyMarkers finds legacy PRESERVE and SYNC markers for migration.
func (p *MarkerParser) ParseLegacyMarkers(content string) []*LegacyMarker {
	var legacy []*LegacyMarker

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	inCodeBlock := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Track code block state
		if codeBlockStartRegex.MatchString(strings.TrimSpace(line)) {
			if inCodeBlock {
				if codeBlockEndRegex.MatchString(strings.TrimSpace(line)) {
					inCodeBlock = false
				}
			} else {
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			continue
		}

		// Check for PRESERVE marker
		if matches := legacyPreserveRegex.FindStringSubmatch(line); matches != nil {
			legacy = append(legacy, &LegacyMarker{
				Type:                LegacyPreserve,
				LineNumber:          lineNum,
				Raw:                 strings.TrimSpace(matches[0]),
				SuggestedRegionName: p.suggestRegionName(matches[1], lineNum),
			})
		}

		// Check for SYNC marker
		if matches := legacySyncRegex.FindStringSubmatch(line); matches != nil {
			legacy = append(legacy, &LegacyMarker{
				Type:                LegacySync,
				LineNumber:          lineNum,
				Raw:                 strings.TrimSpace(matches[0]),
				SuggestedRegionName: p.suggestRegionName(matches[1], lineNum),
			})
		}
	}

	return legacy
}

// suggestRegionName generates a suggested region name from legacy marker content.
func (p *MarkerParser) suggestRegionName(content string, lineNum int) string {
	// Clean up the content
	content = strings.ToLower(strings.TrimSpace(content))

	// Remove common suffixes
	content = strings.TrimSuffix(content, "-owned")
	content = strings.TrimSuffix(content, "satellite")
	content = strings.TrimSuffix(content, "knossos")

	// Convert to kebab-case
	content = strings.ReplaceAll(content, " ", "-")
	content = strings.ReplaceAll(content, "_", "-")

	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, c := range content {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result.WriteRune(c)
		}
	}

	name := result.String()
	if name == "" || name == "-" {
		name = "region-" + itoa(lineNum)
	}

	// Ensure starts with letter
	if name[0] >= '0' && name[0] <= '9' {
		name = "region-" + name
	}

	return name
}

// ValidateRegionName checks if a region name follows naming conventions.
// Region names must be kebab-case: lowercase letters, numbers, and hyphens.
// Must start with a letter.
func ValidateRegionName(name string) error {
	if name == "" {
		return errors.New(errors.CodeUsageError, "region name cannot be empty")
	}

	if len(name) > 64 {
		return errors.New(errors.CodeUsageError, "region name too long (max 64 characters)")
	}

	// Must start with lowercase letter
	if name[0] < 'a' || name[0] > 'z' {
		return errors.New(errors.CodeUsageError, "region name must start with a lowercase letter")
	}

	// Check all characters
	for i, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			continue
		}
		if c == '-' && i > 0 && i < len(name)-1 {
			continue
		}
		return errors.NewWithDetails(errors.CodeUsageError,
			"invalid character in region name",
			map[string]interface{}{
				"character": string(c),
				"position":  i,
				"name":      name,
			})
	}

	// Cannot end with hyphen
	if name[len(name)-1] == '-' {
		return errors.New(errors.CodeUsageError, "region name cannot end with a hyphen")
	}

	// Cannot have consecutive hyphens
	if strings.Contains(name, "--") {
		return errors.New(errors.CodeUsageError, "region name cannot contain consecutive hyphens")
	}

	return nil
}

// BuildMarker constructs a marker string from components.
func BuildMarker(directive MarkerDirective, regionName string, options map[string]string) string {
	var sb strings.Builder
	sb.WriteString("<!-- KNOSSOS:")
	sb.WriteString(string(directive))
	sb.WriteString(" ")
	sb.WriteString(regionName)

	if len(options) > 0 {
		for key, value := range options {
			sb.WriteString(" ")
			sb.WriteString(key)
			sb.WriteString("=")
			sb.WriteString(value)
		}
	}

	sb.WriteString(" -->")
	return sb.String()
}

// BuildStartMarker constructs a START marker string.
func BuildStartMarker(regionName string, options map[string]string) string {
	return BuildMarker(DirectiveStart, regionName, options)
}

// BuildEndMarker constructs an END marker string.
func BuildEndMarker(regionName string) string {
	return BuildMarker(DirectiveEnd, regionName, nil)
}

// BuildAnchorMarker constructs an ANCHOR marker string.
func BuildAnchorMarker(regionName string) string {
	return BuildMarker(DirectiveAnchor, regionName, nil)
}

// WrapContent wraps content with START and END markers.
func WrapContent(regionName string, content string, options map[string]string) string {
	var sb strings.Builder
	sb.WriteString(BuildStartMarker(regionName, options))
	sb.WriteString("\n")
	sb.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString(BuildEndMarker(regionName))
	return sb.String()
}

// itoa converts an integer to a string without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}

	negative := i < 0
	if negative {
		i = -i
	}

	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}
