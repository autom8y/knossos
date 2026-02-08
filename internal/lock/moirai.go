package lock

import (
	"encoding/json"
	"os"
	"time"
)

// MoiraiLock represents the Moirai agent's lock file structure.
// This is distinct from LockMetadata (advisory session locks).
type MoiraiLock struct {
	Agent             string    `json:"agent"`
	AcquiredAt        time.Time `json:"acquired_at"`
	SessionID         string    `json:"session_id"`
	StaleAfterSeconds int       `json:"stale_after_seconds"`
}

// ReadMoiraiLock reads and parses a Moirai lock file at the given path.
// Returns the parsed lock or an error if the file cannot be read or parsed.
func ReadMoiraiLock(lockPath string) (*MoiraiLock, error) {
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, err
	}

	var lock MoiraiLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, err
	}

	return &lock, nil
}

// IsMoiraiLockStale returns true if the lock has exceeded its stale threshold.
func IsMoiraiLockStale(lock *MoiraiLock) bool {
	age := time.Since(lock.AcquiredAt)
	threshold := time.Duration(lock.StaleAfterSeconds) * time.Second
	return age > threshold
}
