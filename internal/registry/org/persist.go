package org

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/fileutil"
)

// catalogFileName is the canonical name for the domains.yaml catalog file.
const catalogFileName = "domains.yaml"

// CatalogPath returns the absolute path for the catalog file given an org context.
func CatalogPath(ctx OrgContext) string {
	return filepath.Join(ctx.RegistryDir(), catalogFileName)
}

// LoadCatalog reads and parses a DomainCatalog from path.
// Returns an error if the file is missing or cannot be parsed.
func LoadCatalog(path string) (*DomainCatalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("catalog not found at %s: run 'ari registry sync' first", path)
		}
		return nil, fmt.Errorf("read catalog %s: %w", path, err)
	}

	var catalog DomainCatalog
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("parse catalog %s: %w", path, err)
	}

	return &catalog, nil
}

// SaveCatalog persists a DomainCatalog to path using an atomic write.
// Creates parent directories as needed.
func SaveCatalog(path string, catalog *DomainCatalog) error {
	data, err := yaml.Marshal(catalog)
	if err != nil {
		return fmt.Errorf("marshal catalog: %w", err)
	}

	if err := fileutil.AtomicWriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write catalog %s: %w", path, err)
	}

	return nil
}
