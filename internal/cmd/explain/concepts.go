package explain

import (
	"github.com/autom8y/knossos/internal/concept"
)

// ConceptEntry is re-exported from the concept package for backward compatibility.
type ConceptEntry = concept.ConceptEntry

// LookupConcept delegates to the concept package.
func LookupConcept(input string) (*ConceptEntry, error) {
	return concept.LookupConcept(input)
}

// AllConcepts delegates to the concept package.
func AllConcepts() []*ConceptEntry {
	return concept.AllConcepts()
}
