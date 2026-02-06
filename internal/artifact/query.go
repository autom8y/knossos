package artifact

// QueryFilter specifies filter criteria for artifact queries.
type QueryFilter struct {
	Phase      Phase        `json:"phase,omitempty"`
	Type       ArtifactType `json:"type,omitempty"`
	Specialist string       `json:"specialist,omitempty"`
	SessionID  string       `json:"session_id,omitempty"`
}

// QueryResult contains the results of an artifact query.
type QueryResult struct {
	Entries []Entry     `json:"entries"`
	Count   int         `json:"count"`
	Filter  QueryFilter `json:"filter"`
}

// Querier provides query operations on the artifact registry.
type Querier struct {
	registry *Registry
}

// NewQuerier creates a new Querier.
func NewQuerier(registry *Registry) *Querier {
	return &Querier{registry: registry}
}

// Query executes a filtered query against the project registry.
// Multiple filter fields are ANDed together.
func (q *Querier) Query(filter QueryFilter) (*QueryResult, error) {
	projectReg, err := q.registry.LoadProjectRegistry()
	if err != nil {
		return nil, err
	}

	var matches []Entry

	for _, entry := range projectReg.Artifacts {
		if !q.matchesFilter(entry, filter) {
			continue
		}
		matches = append(matches, entry)
	}

	return &QueryResult{
		Entries: matches,
		Count:   len(matches),
		Filter:  filter,
	}, nil
}

// matchesFilter checks if an entry matches all non-empty filter criteria.
func (q *Querier) matchesFilter(entry Entry, filter QueryFilter) bool {
	if filter.Phase != "" && entry.Phase != filter.Phase {
		return false
	}
	if filter.Type != "" && entry.ArtifactType != filter.Type {
		return false
	}
	if filter.Specialist != "" && entry.Specialist != filter.Specialist {
		return false
	}
	if filter.SessionID != "" && entry.SessionID != filter.SessionID {
		return false
	}
	return true
}

// ListPhases returns all phases with their artifact counts.
func (q *Querier) ListPhases() (map[Phase]int, error) {
	projectReg, err := q.registry.LoadProjectRegistry()
	if err != nil {
		return nil, err
	}

	counts := make(map[Phase]int)
	for phase, ids := range projectReg.Indexes.ByPhase {
		counts[phase] = len(ids)
	}
	return counts, nil
}

// ListTypes returns all types with their artifact counts.
func (q *Querier) ListTypes() (map[ArtifactType]int, error) {
	projectReg, err := q.registry.LoadProjectRegistry()
	if err != nil {
		return nil, err
	}

	counts := make(map[ArtifactType]int)
	for t, ids := range projectReg.Indexes.ByType {
		counts[t] = len(ids)
	}
	return counts, nil
}

// ListSpecialists returns all specialists with their artifact counts.
func (q *Querier) ListSpecialists() (map[string]int, error) {
	projectReg, err := q.registry.LoadProjectRegistry()
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for s, ids := range projectReg.Indexes.BySpecialist {
		counts[s] = len(ids)
	}
	return counts, nil
}

// ListSessions returns all sessions with their artifact counts.
func (q *Querier) ListSessions() (map[string]int, error) {
	projectReg, err := q.registry.LoadProjectRegistry()
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for s, ids := range projectReg.Indexes.BySession {
		counts[s] = len(ids)
	}
	return counts, nil
}
