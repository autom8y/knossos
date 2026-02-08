package frontmatter

import "errors"

// Sentinel errors for frontmatter parsing.
var (
	// ErrMissingOpenDelimiter indicates content does not start with "---\n".
	ErrMissingOpenDelimiter = errors.New("missing frontmatter opening delimiter")

	// ErrMissingCloseDelimiter indicates no closing "---\n" delimiter was found.
	ErrMissingCloseDelimiter = errors.New("missing frontmatter closing delimiter")
)
