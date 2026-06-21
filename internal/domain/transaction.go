package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Transaction is the transport-agnostic transaction model. Money is
// decimal.Decimal (lossless on the REST path); only the gRPC adapter converts
// to/from double. Joined display names (category/account/creator) are not part
// of the domain — each adapter enriches its own response as needed.
type Transaction struct {
	ID           uuid.UUID
	Type         string // income | expense
	Amount       decimal.Decimal
	Currency     string
	AmountBase   decimal.Decimal
	BaseCurrency string // "RSD" (MVP)
	AccountID    uuid.UUID
	CategoryID   uuid.UUID
	Description  string
	Date         time.Time
	CreatedBy    uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
