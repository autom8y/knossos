package materialize

// SyncRite satisfies the rite.Syncer interface.
func (m *Materializer) SyncRite(riteName string, keepOrphans bool) error {
	_, err := m.Sync(SyncOptions{Scope: ScopeRite, RiteName: riteName, KeepOrphans: keepOrphans})
	return err
}
