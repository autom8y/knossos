package rite

// Syncer can sync rite resources. Implemented by materialize.Materializer.
// This interface exists to break the upward dependency from rite -> materialize.
type Syncer interface {
	// SyncRite performs a rite-scope sync for the named rite.
	SyncRite(riteName string, keepOrphans bool) error
}
