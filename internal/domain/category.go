package domain

import (
	"time"

	"github.com/google/uuid"
)

// Category is the transport-agnostic category model. ParentID is nil for a
// root (top-level) category. No pgtype/family_id/deleted_by leak into the domain.
type Category struct {
	ID        uuid.UUID
	Name      string
	Type      string // income | expense
	ParentID  *uuid.UUID
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
