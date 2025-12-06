package dto

import (
	"time"

	"github.com/google/uuid"
)

// Category Requests -->

// CreateCategoryRequest represents data needed to create a new category
type CreateCategoryRequest struct {
	Name     string     `json:"name" validate:"required,min=1,max=100" example:"Groceries"`
	Type     string     `json:"type" validate:"required,oneof=income expense" example:"expense"`
	ParentID *uuid.UUID `json:"parent_id,omitempty" example:"e3b5cae7-cf1d-4c22-a0fe-4014590a7a84"`
}

// UpdateCategoryRequest represents data for updating a category (partial update)
type UpdateCategoryRequest struct {
	Name     *string    `json:"name,omitempty" validate:"omitempty,min=1,max=100" example:"Updated Category Name"`
	ParentID *uuid.UUID `json:"parent_id,omitempty" example:"e3b5cae7-cf1d-4c22-a0fe-4014590a7a84"`
	IsActive *bool      `json:"is_active,omitempty" example:"true"`
}

// Category Responses <--

// CategoryResponse represents category information
type CategoryResponse struct {
	ID        uuid.UUID  `json:"id" example:"e3b5cae7-cf1d-4c22-a0fe-4014590a7a85"`
	Name      string     `json:"name" example:"Groceries"`
	Type      string     `json:"type" example:"expense"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty" example:"e3b5cae7-cf1d-4c22-a0fe-4014590a7a84"`
	IsActive  bool       `json:"is_active" example:"true"`
	CreatedAt time.Time  `json:"created_at" example:"2025-11-01T10:00:00Z"`
	UpdatedAt time.Time  `json:"updated_at" example:"2025-12-06T18:30:00Z"`
}

// CategoryListResponse represents a list of categories
type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
}
