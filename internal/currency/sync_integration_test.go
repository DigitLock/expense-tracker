package currency

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/testutil"
)

type rateRow struct {
	from, to, source string
	rate             string
	fetchedAt        time.Time
}

func dumpRates(t *testing.T, pool *pgxpool.Pool) []rateRow {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT from_currency, to_currency, rate::text, source, fetched_at
		 FROM exchange_rates WHERE to_currency = 'RSD' ORDER BY from_currency`)
	if err != nil {
		t.Fatalf("query rates: %v", err)
	}
	defer rows.Close()

	var out []rateRow
	for rows.Next() {
		var r rateRow
		if err := rows.Scan(&r.from, &r.to, &r.rate, &r.source, &r.fetchedAt); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, r)
	}
	return out
}

// TestSync_Integration exercises Sync against the real repository and test DB
// (skipped when TEST_DATABASE_URL is unset). Requires migration 012 (source,
// fetched_at, USD-widened CHECK), which `make test` applies via test-setup.
func TestSync_Integration(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()

	if _, err := pool.Exec(ctx, "TRUNCATE exchange_rates"); err != nil {
		t.Fatalf("truncate exchange_rates: %v", err)
	}

	repo := repository.New(pool).ExchangeRates
	fetcher := &fakeFetcher{rates: []FetchedRate{
		crsRate("EUR", 0.0085228051),
		crsRate("USD", 0.0098699845),
	}}
	syncer := NewSyncer(fetcher, repo)

	// First sync: two inverted rows persisted.
	if _, err := syncer.Sync(ctx); err != nil {
		t.Fatalf("sync run1: %v", err)
	}
	rows := dumpRates(t, pool)
	if len(rows) != 2 {
		t.Fatalf("rows after run1 = %d, want 2", len(rows))
	}

	want := map[string]string{"EUR": "117.332262", "USD": "101.317282"}
	for _, r := range rows {
		if r.to != "RSD" {
			t.Errorf("%s: to = %q, want RSD", r.from, r.to)
		}
		if r.rate != want[r.from] {
			t.Errorf("%s->RSD rate = %s, want %s", r.from, r.rate, want[r.from])
		}
		if r.source != "fawazahmed0" {
			t.Errorf("%s: source = %q, want fawazahmed0", r.from, r.source)
		}
		if r.fetchedAt.IsZero() {
			t.Errorf("%s: fetched_at is zero", r.from)
		}
	}
	firstFetched := rows[0].fetchedAt

	// Second sync on the same date: upsert (update), not a duplicate.
	time.Sleep(2 * time.Millisecond)
	if _, err := syncer.Sync(ctx); err != nil {
		t.Fatalf("sync run2: %v", err)
	}
	rows2 := dumpRates(t, pool)
	if len(rows2) != 2 {
		t.Fatalf("rows after run2 = %d, want 2 (idempotent upsert, no duplicates)", len(rows2))
	}
	if !rows2[0].fetchedAt.After(firstFetched) {
		t.Errorf("fetched_at not advanced on re-sync: %s -> %s", firstFetched, rows2[0].fetchedAt)
	}
}
