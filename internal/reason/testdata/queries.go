// Package testdata contains the test corpus for the reasoning pipeline.
package testdata

// TestQuery defines a single test case for the reasoning pipeline.
type TestQuery struct {
	// ID is a unique identifier for the test case.
	ID string

	// Query is the user's question.
	Query string

	// ExpectedTier is the expected confidence tier ("HIGH", "MEDIUM", "LOW").
	ExpectedTier string

	// ExpectedIntent is the expected action tier ("OBSERVE", "RECORD", "ACT").
	ExpectedIntent string

	// ExpectedDomains are the expected domain hints from intent classification.
	ExpectedDomains []string

	// ExpectedAnswerable indicates whether the query should be answerable.
	ExpectedAnswerable bool

	// Description explains the test scenario.
	Description string
}

// HighConfidenceQueries are test queries expected to produce HIGH confidence responses.
var HighConfidenceQueries = []TestQuery{
	{ID: "H-01", Query: "How does the sync pipeline work?", ExpectedTier: "HIGH", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"architecture"}, ExpectedAnswerable: true, Description: "Core architectural query; recent .know/"},
	{ID: "H-02", Query: "What error handling patterns does knossos use?", ExpectedTier: "HIGH", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"conventions"}, ExpectedAnswerable: true, Description: "Convention query; well-covered domain"},
	{ID: "H-03", Query: "Explain the session lifecycle state machine", ExpectedTier: "HIGH", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"architecture"}, ExpectedAnswerable: true, Description: "Specific subsystem query"},
	{ID: "H-04", Query: "What is the materializer pattern?", ExpectedTier: "HIGH", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"architecture", "conventions"}, ExpectedAnswerable: true, Description: "Multi-domain query"},
	{ID: "H-05", Query: "How does the provenance manifest track file ownership?", ExpectedTier: "HIGH", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"architecture"}, ExpectedAnswerable: true, Description: "Deep infrastructure query"},
	{ID: "H-06", Query: "What testing patterns are used in the codebase?", ExpectedTier: "HIGH", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"test-coverage", "conventions"}, ExpectedAnswerable: true, Description: "Cross-domain with good coverage"},
	{ID: "H-07", Query: "Describe the hook event pipeline from CC to ari", ExpectedTier: "HIGH", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"architecture"}, ExpectedAnswerable: true, Description: "Integration flow query"},
	{ID: "H-08", Query: "What are the naming conventions for Go packages?", ExpectedTier: "HIGH", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"conventions"}, ExpectedAnswerable: true, Description: "Specific convention query"},
}

// MediumConfidenceQueries are test queries expected to produce MEDIUM confidence responses.
var MediumConfidenceQueries = []TestQuery{
	{ID: "M-01", Query: "What test coverage gaps exist?", ExpectedTier: "MEDIUM", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"test-coverage"}, ExpectedAnswerable: true, Description: "test-coverage has 5-day half-life; frequently stale"},
	{ID: "M-02", Query: "What was in the latest release?", ExpectedTier: "MEDIUM", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"release"}, ExpectedAnswerable: true, Description: "release has 3-day half-life; ages fast"},
	{ID: "M-03", Query: "How do features interact with the rite system?", ExpectedTier: "MEDIUM", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"feat/", "architecture"}, ExpectedAnswerable: true, Description: "feat/ domain may be partially stale"},
	{ID: "M-04", Query: "What are the design constraints for the search package?", ExpectedTier: "MEDIUM", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"design-constraints"}, ExpectedAnswerable: true, Description: "Specific but potentially aged"},
	{ID: "M-05", Query: "How does cross-rite coordination work with processions?", ExpectedTier: "MEDIUM", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"architecture"}, ExpectedAnswerable: true, Description: "Complex subsystem, partial coverage"},
	{ID: "M-06", Query: "What scar tissue exists around the session module?", ExpectedTier: "MEDIUM", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"scar-tissue"}, ExpectedAnswerable: true, Description: "scar-tissue has 10-day half-life"},
}

// LowConfidenceQueries are test queries expected to produce LOW confidence responses.
var LowConfidenceQueries = []TestQuery{
	{ID: "L-01", Query: "How does the Kubernetes migration work?", ExpectedTier: "LOW", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"kubernetes-migration"}, ExpectedAnswerable: true, Description: "Unregistered domain; zero coverage"},
	{ID: "L-02", Query: "What is the authentication system architecture?", ExpectedTier: "LOW", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"authentication"}, ExpectedAnswerable: true, Description: "Unregistered domain"},
	{ID: "L-03", Query: "How does billing work?", ExpectedTier: "LOW", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"billing"}, ExpectedAnswerable: true, Description: "Unregistered domain"},
	{ID: "L-04", Query: "What CI/CD pipeline does the mobile app use?", ExpectedTier: "LOW", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"ci-cd"}, ExpectedAnswerable: true, Description: "Unregistered domains"},
	{ID: "L-05", Query: "How does the frontend React code work?", ExpectedTier: "LOW", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"frontend"}, ExpectedAnswerable: true, Description: "Unregistered domain"},
	{ID: "L-06", Query: "What is the database schema for user accounts?", ExpectedTier: "LOW", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"database"}, ExpectedAnswerable: true, Description: "Unregistered domain"},
}

// IntentEdgeCaseQueries are test queries for intent classification edge cases.
var IntentEdgeCaseQueries = []TestQuery{
	{ID: "I-01", Query: "Update the scar tissue for the session bug", ExpectedTier: "", ExpectedIntent: "RECORD", ExpectedDomains: []string{"scar-tissue"}, ExpectedAnswerable: false, Description: "Record verb 'update' + domain hint"},
	{ID: "I-02", Query: "Run a hotfix for the login issue", ExpectedTier: "", ExpectedIntent: "ACT", ExpectedDomains: nil, ExpectedAnswerable: false, Description: "Act verb 'run' + 'hotfix'"},
	{ID: "I-03", Query: "How do I deploy a new version?", ExpectedTier: "", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"release"}, ExpectedAnswerable: true, Description: "'How do I' guard clause overrides 'deploy'"},
	{ID: "I-04", Query: "Create a new knowledge file for testing", ExpectedTier: "", ExpectedIntent: "RECORD", ExpectedDomains: []string{"test-coverage"}, ExpectedAnswerable: false, Description: "Record verb 'create'"},
	{ID: "I-05", Query: "What does the deploy script do?", ExpectedTier: "", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"release"}, ExpectedAnswerable: true, Description: "'What does' guard clause overrides 'deploy'"},
	{ID: "I-06", Query: "Tell me about the release process", ExpectedTier: "", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"release"}, ExpectedAnswerable: true, Description: "'Tell me about' guard clause"},
	{ID: "I-07", Query: "Execute the migration script in prod", ExpectedTier: "", ExpectedIntent: "ACT", ExpectedDomains: nil, ExpectedAnswerable: false, Description: "Act verb 'execute'"},
	{ID: "I-08", Query: "Explain how to run integration tests", ExpectedTier: "", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"test-coverage"}, ExpectedAnswerable: true, Description: "'Explain how to' guard clause overrides 'run'"},
	{ID: "I-09", Query: "File a scar for the OOM incident", ExpectedTier: "", ExpectedIntent: "RECORD", ExpectedDomains: []string{"scar-tissue"}, ExpectedAnswerable: false, Description: "Record verb 'file' + 'scar' domain"},
	{ID: "I-10", Query: "Show me the deployment architecture", ExpectedTier: "", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"architecture"}, ExpectedAnswerable: true, Description: "'Show me' guard clause overrides 'deployment'"},
	{ID: "I-11", Query: "Revert the last commit on main", ExpectedTier: "", ExpectedIntent: "ACT", ExpectedDomains: nil, ExpectedAnswerable: false, Description: "Act verb 'revert'"},
	{ID: "I-12", Query: "What are the best practices for error handling?", ExpectedTier: "", ExpectedIntent: "OBSERVE", ExpectedDomains: []string{"conventions"}, ExpectedAnswerable: true, Description: "Pure knowledge query"},
}

// AllQueries returns the combined test corpus for pipeline integration tests.
func AllQueries() []TestQuery {
	var all []TestQuery
	all = append(all, HighConfidenceQueries...)
	all = append(all, MediumConfidenceQueries...)
	all = append(all, LowConfidenceQueries...)
	return all
}
