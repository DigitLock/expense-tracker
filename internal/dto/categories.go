package dto

import (
	"time"

	"github.com/google/uuid"
)

// --- Requests ---

// CreateCategoryRequest - запрос на создание категории
type CreateCategoryRequest struct {
	Name     string     `json:"name" validate:"required,min=1,max=100"`
	Type     string     `json:"type" validate:"required,oneof=income expense"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

// UpdateCategoryRequest - запрос на обновление категории (partial)
type UpdateCategoryRequest struct {
	Name     *string    `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	IsActive *bool      `json:"is_active,omitempty"`
}

// --- Responses ---

// CategoryResponse - категория в ответе API
type CategoryResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CategoryListResponse - список категорий
type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
}
