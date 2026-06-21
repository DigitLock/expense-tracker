package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Account is the transport-agnostic account model. It carries clean Go types
// (no pgtype, no family_id/deleted_by) so REST and gRPC adapters map from it
// directly. Money is decimal.Decimal (same type as sqlc/REST-DTO) to keep the
// REST path lossless; only the gRPC adapter converts to/from double.
type Account struct {
	ID             uuid.UUID
	Name           string
	Type           string
	Currency       string
	InitialBalance decimal.Decimal
	CurrentBalance decimal.Decimal
	Description    string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
