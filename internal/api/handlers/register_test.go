package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DigitLock/expense-tracker/internal/api/handlers"
	"github.com/DigitLock/expense-tracker/internal/auth"
	"github.com/DigitLock/expense-tracker/internal/dto"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/testutil"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newAuthHandler(t *testing.T) (*handlers.AuthHandler, *pgxpool.Pool) {
	t.Helper()
	pool := testutil.Pool(t)
	testutil.Truncate(t, pool)
	repos := repository.New(pool)
	jwtService := auth.NewJWTService("test-secret", 24)
	return handlers.NewAuthHandler(repos.Users, jwtService), pool
}

func doRegister(h *handlers.AuthHandler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Register(w, req)
	return w
}

func TestAuthHandler_Register_Success(t *testing.T) {
	h, pool := newAuthHandler(t)

	w := doRegister(h, `{"email":"new@example.com","password":"Secret123","name":"Igor"}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool             `json:"success"`
		Data    dto.LoginResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v; body: %s", err, w.Body.String())
	}

	if !resp.Success {
		t.Error("success = false, want true")
	}
	if resp.Data.Token == "" {
		t.Error("token is empty")
	}
	if resp.Data.User.Role != "owner" {
		t.Errorf("user.role = %q, want \"owner\"", resp.Data.User.Role)
	}
	if resp.Data.User.FamilyID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Error("user.family_id is the zero UUID")
	}

	// Default family name applied: "{name}'s family".
	var familyName string
	if err := pool.QueryRow(context.Background(),
		"SELECT name FROM families WHERE id = $1", resp.Data.User.FamilyID,
	).Scan(&familyName); err != nil {
		t.Fatalf("query family name: %v", err)
	}
	if want := "Igor's family"; familyName != want {
		t.Errorf("family.name = %q, want %q", familyName, want)
	}
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	h, _ := newAuthHandler(t)

	body := `{"email":"dup@example.com","password":"Secret123","name":"First"}`
	if w := doRegister(h, body); w.Code != http.StatusCreated {
		t.Fatalf("first register status = %d, want 201; body: %s", w.Code, w.Body.String())
	}

	w := doRegister(h, `{"email":"dup@example.com","password":"Other123","name":"Second"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body: %s", w.Code, w.Body.String())
	}

	var resp dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if resp.Error.Code != "EMAIL_ALREADY_EXISTS" {
		t.Errorf("error.code = %q, want \"EMAIL_ALREADY_EXISTS\"", resp.Error.Code)
	}
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	h, _ := newAuthHandler(t)

	// password too short (<8) and name missing.
	w := doRegister(h, `{"email":"bad@example.com","password":"short"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", w.Code, w.Body.String())
	}

	var resp dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("error.code = %q, want \"VALIDATION_ERROR\"", resp.Error.Code)
	}
	if len(resp.Error.Details) == 0 {
		t.Fatal("error.details is empty, want field-level validation errors")
	}

	// Field names are reported as json names (RegisterTagNameFunc), not Go
	// struct field names: "password"/"name", not "Password"/"Name".
	fields := map[string]bool{}
	for _, d := range resp.Error.Details {
		fields[d.Field] = true
	}
	for _, want := range []string{"password", "name"} {
		if !fields[want] {
			t.Errorf("validation details missing field %q; got fields %v", want, fields)
		}
	}
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	h, _ := newAuthHandler(t)

	w := doRegister(h, `{"email":`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", w.Code, w.Body.String())
	}

	var resp dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if resp.Error.Code != "INVALID_JSON" {
		t.Errorf("error.code = %q, want \"INVALID_JSON\"", resp.Error.Code)
	}
}
