package dto

import "time"

// SuccessResponse is a generic successful API response wrapper
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data,omitempty"`
}

// MessageResponse represents a simple success message response
type MessageResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Operation completed successfully"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Success bool        `json:"success" example:"false"`
	Error   ErrorDetail `json:"error"`
}

// ErrorDetail contains detailed error information
type ErrorDetail struct {
	Code    string            `json:"code" example:"VALIDATION_ERROR"`
	Message string            `json:"message" example:"Invalid input"`
	Details []ValidationError `json:"details,omitempty"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"Invalid email format"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page       int `json:"page" example:"1"`
	PerPage    int `json:"per_page" example:"50"`
	Total      int `json:"total" example:"150"`
	TotalPages int `json:"total_pages" example:"3"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string    `json:"status" example:"healthy"`
	Timestamp time.Time `json:"timestamp" example:"2025-12-06T18:30:00Z"`
	Version   string    `json:"version" example:"1.0.0"`
	Database  string    `json:"database" example:"connected"`
}

// NewSuccessResponse creates a successful API response with data
func NewSuccessResponse(data interface{}) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Data:    data,
	}
}

// NewMessageResponse creates a successful API response with a message
func NewMessageResponse(message string) MessageResponse {
	return MessageResponse{
		Success: true,
		Message: message,
	}
}

// NewErrorResponse creates an error API response with code, message and optional validation details
func NewErrorResponse(code, message string, details []ValidationError) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
