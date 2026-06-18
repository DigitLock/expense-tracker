package currency

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

// fakeFetcher is a test RateFetcher returning canned rates / error.
type fakeFetcher struct {
	rates []FetchedRate
	err   error
}

func (f *fakeFetcher) FetchRates(_ context.Context) ([]FetchedRate, error) {
	return f.rates, f.err
}

// fakeUpserter is a test rateUpserter recording every call.
type fakeUpserter struct {
	calls []repository.CreateRateInput
	err   error
}

func (u *fakeUpserter) Upsert(_ context.Context, input repository.CreateRateInput) (sqlc.ExchangeRate, error) {
	u.calls = append(u.calls, input)
	return sqlc.ExchangeRate{}, u.err
}

// crsRate builds a CRS-shaped rate (base=RSD -> target).
func crsRate(target string, rate float64) FetchedRate {
	return FetchedRate{
		From:           "RSD",
		To:             target,
		Rate:           rate,
		UpdatedAt:      time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC),
		SourceProvider: "fawazahmed0",
	}
}

func TestSync_MappingInversion(t *testing.T) {
	fetcher := &fakeFetcher{rates: []FetchedRate{
		crsRate("EUR", 0.0085228051),
		crsRate("USD", 0.0098699845),
	}}
	upserter := &fakeUpserter{}
	syncer := NewSyncer(fetcher, upserter)

	result, err := syncer.Sync(context.Background())
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if result.SyncedPairs != 2 {
		t.Errorf("SyncedPairs = %d, want 2", result.SyncedPairs)
	}
	if len(result.Missing) != 0 {
		t.Errorf("Missing = %v, want empty", result.Missing)
	}
	if len(upserter.calls) != 2 {
		t.Fatalf("upsert calls = %d, want 2", len(upserter.calls))
	}

	want := map[string]string{"EUR": "117.332262", "USD": "101.317282"}
	for _, c := range upserter.calls {
		if c.ToCurrency != "RSD" {
			t.Errorf("%s: ToCurrency = %q, want RSD", c.FromCurrency, c.ToCurrency)
		}
		if c.Source != "fawazahmed0" {
			t.Errorf("%s: Source = %q, want fawazahmed0", c.FromCurrency, c.Source)
		}
		if c.FetchedAt.IsZero() {
			t.Errorf("%s: FetchedAt is zero, want non-empty", c.FromCurrency)
		}
		wantRate, ok := want[c.FromCurrency]
		if !ok {
			t.Errorf("unexpected upsert for %s", c.FromCurrency)
			continue
		}
		if !c.Rate.Equal(decimal.RequireFromString(wantRate)) {
			t.Errorf("%s->RSD rate = %s, want %s", c.FromCurrency, c.Rate, wantRate)
		}
	}
}

func TestSync_PartialResponse(t *testing.T) {
	// CRS returns only EUR — USD must be reported missing, not written.
	fetcher := &fakeFetcher{rates: []FetchedRate{crsRate("EUR", 0.0085228051)}}
	upserter := &fakeUpserter{}

	result, err := NewSyncer(fetcher, upserter).Sync(context.Background())
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if result.SyncedPairs != 1 {
		t.Errorf("SyncedPairs = %d, want 1", result.SyncedPairs)
	}
	if len(upserter.calls) != 1 || upserter.calls[0].FromCurrency != "EUR" {
		t.Errorf("upsert calls = %+v, want single EUR", upserter.calls)
	}
	if len(result.Missing) != 1 || result.Missing[0] != "USD" {
		t.Errorf("Missing = %v, want [USD]", result.Missing)
	}
}

func TestSync_ZeroRateSkipped(t *testing.T) {
	// EUR has rate 0 (cannot invert) -> skipped; USD valid -> written.
	fetcher := &fakeFetcher{rates: []FetchedRate{
		crsRate("EUR", 0),
		crsRate("USD", 0.0098699845),
	}}
	upserter := &fakeUpserter{}

	result, err := NewSyncer(fetcher, upserter).Sync(context.Background())
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if result.SyncedPairs != 1 {
		t.Errorf("SyncedPairs = %d, want 1", result.SyncedPairs)
	}
	if len(upserter.calls) != 1 || upserter.calls[0].FromCurrency != "USD" {
		t.Errorf("upsert calls = %+v, want single USD (EUR skipped)", upserter.calls)
	}
}

func TestSync_OutdatedStillWritten(t *testing.T) {
	r := crsRate("EUR", 0.0085228051)
	r.IsOutdated = true
	upserter := &fakeUpserter{}

	result, err := NewSyncer(&fakeFetcher{rates: []FetchedRate{r}}, upserter).Sync(context.Background())
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if result.SyncedPairs != 1 || len(upserter.calls) != 1 {
		t.Errorf("outdated rate should still be upserted: synced=%d calls=%d", result.SyncedPairs, len(upserter.calls))
	}
}

func TestSync_EmptyResponse(t *testing.T) {
	upserter := &fakeUpserter{}

	result, err := NewSyncer(&fakeFetcher{rates: nil}, upserter).Sync(context.Background())
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if result.SyncedPairs != 0 {
		t.Errorf("SyncedPairs = %d, want 0", result.SyncedPairs)
	}
	if len(upserter.calls) != 0 {
		t.Errorf("upsert calls = %d, want 0", len(upserter.calls))
	}
	if len(result.Missing) != 2 {
		t.Errorf("Missing = %v, want [EUR USD]", result.Missing)
	}
}

func TestSync_FetchError(t *testing.T) {
	upserter := &fakeUpserter{}
	fetcher := &fakeFetcher{err: errors.New("connection refused")}

	_, err := NewSyncer(fetcher, upserter).Sync(context.Background())
	if err == nil {
		t.Fatal("Sync error = nil, want error")
	}
	if len(upserter.calls) != 0 {
		t.Errorf("upsert called %d times on fetch error, want 0 (DB untouched)", len(upserter.calls))
	}
}
