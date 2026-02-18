// Package materialize re-exports collision types from the userscope sub-package.
package materialize

import (
	"github.com/autom8y/knossos/internal/materialize/userscope"
)

// Type alias for backward compatibility.
type CollisionChecker = userscope.CollisionChecker

// Re-export constructor.
var NewCollisionChecker = userscope.NewCollisionChecker
