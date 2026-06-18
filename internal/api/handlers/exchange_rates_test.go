package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DigitLock/expense-tracker/internal/api/handlers"
	"github.com/DigitLock/expense-tracker/internal/currency"
	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
	"github.com/DigitLock/expense-tracker/internal/dto"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

// fakeFetcher / fakeUpserter let us build a real currency.Syncer without a CRS
// or DB. fakeUpserter's Upsert signature matches the seam interface exactly.
type fakeFetcher struct {
	rates []currency.FetchedRate
	err   error
}

func (f *fakeFetcher) FetchRates(_ context.Context) ([]currency.FetchedRate, error) {
	return f.rates, f.err
}

type fakeUpserter struct{}

func (fakeUpserter) Upsert(_ context.Context, _ repository.CreateRateInput) (sqlc.ExchangeRate, error) {
	return sqlc.ExchangeRate{}, nil
}

func postSync(h *handlers.ExchangeRateHandler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/exchange-rates/sync", nil)
	w := httptest.NewRecorder()
	h.Sync(w, req)
	return w
}

func TestExchangeRateHandler_Sync_Success(t *testing.T) {
	fetcher := &fakeFetcher{rates: []currency.FetchedRate{
		{From: "RSD", To: "EUR", Rate: 0.0085228051, SourceProvider: "fawazahmed0", UpdatedAt: time.Now().UTC()},
		{From: "RSD", To: "USD", Rate: 0.0098699845, SourceProvider: "fawazahmed0", UpdatedAt: time.Now().UTC()},
	}}
	syncer := currency.NewSyncer(fetcher, fakeUpserter{})
	h := handlers.NewExchangeRateHandler(syncer)

	w := postSync(h)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool                         `json:"success"`
		Data    dto.ExchangeRateSyncResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v; body: %s", err, w.Body.String())
	}
	if !resp.Success {
		t.Error("success = false, want true")
	}
	if resp.Data.SyncedPairs != 2 {
		t.Errorf("synced_pairs = %d, want 2", resp.Data.SyncedPairs)
	}
	if resp.Data.Source != "fawazahmed0" {
		t.Errorf("source = %q, want fawazahmed0", resp.Data.Source)
	}
	if resp.Data.FetchedAt.IsZero() {
		t.Error("fetched_at is zero, want non-empty")
	}
}

func TestExchangeRateHandler_Sync_FetchError(t *testing.T) {
	fetcher := &fakeFetcher{err: context.DeadlineExceeded}
	h := handlers.NewExchangeRateHandler(currency.NewSyncer(fetcher, fakeUpserter{}))

	w := postSync(h)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body: %s", w.Code, w.Body.String())
	}
	assertErrorCode(t, w.Body.Bytes(), "CURRENCY_SERVICE_UNAVAILABLE")
}

func TestExchangeRateHandler_Sync_NilSyncer(t *testing.T) {
	h := handlers.NewExchangeRateHandler(nil)

	w := postSync(h)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body: %s", w.Code, w.Body.String())
	}
	assertErrorCode(t, w.Body.Bytes(), "CURRENCY_SERVICE_UNAVAILABLE")
}

func assertErrorCode(t *testing.T, body []byte, want string) {
	t.Helper()
	var resp dto.ErrorResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if resp.Error.Code != want {
		t.Errorf("error.code = %q, want %q", resp.Error.Code, want)
	}
}
