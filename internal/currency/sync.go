package currency

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

// RateFetcher fetches current rates from the upstream service. *Client satisfies it.
type RateFetcher interface {
	FetchRates(ctx context.Context) ([]FetchedRate, error)
}

// rateUpserter persists a rate row. Signature matches
// (*repository.ExchangeRateRepository).Upsert exactly.
type rateUpserter interface {
	Upsert(ctx context.Context, input repository.CreateRateInput) (sqlc.ExchangeRate, error)
}

// targetCurrencies are the foreign currencies we sync against RSD. CRS is
// queried with base=RSD; each returned pair is inverted into ET's convention
// (rate = RSD per 1 foreign unit, stored as from=foreign, to=RSD).
var targetCurrencies = []string{"EUR", "USD"}

// SyncedPair describes one rate written to ET (in ET's convention, post-inversion).
type SyncedPair struct {
	From   string  // foreign currency (e.g. EUR)
	To     string  // always "RSD"
	Rate   float64 // RSD per 1 unit of From
	Source string
}

// SyncResult summarizes a sync run so callers (st5 trigger handler) can report it.
type SyncResult struct {
	SyncedPairs int
	Pairs       []SyncedPair
	Missing     []string // requested targets CRS did not return
	FetchedAt   time.Time
}

// Syncer fetches rates from CRS and upserts them into ET's exchange_rates,
// applying the inversion into ET's currency convention. It depends on the
// RateFetcher and rateUpserter interfaces (satisfied by *Client and
// *repository.ExchangeRateRepository) so the core can be unit-tested.
type Syncer struct {
	fetcher RateFetcher
	repo    rateUpserter
}

// NewSyncer wires the rate fetcher and the exchange-rate upserter.
func NewSyncer(fetcher RateFetcher, repo rateUpserter) *Syncer {
	return &Syncer{fetcher: fetcher, repo: repo}
}

// Sync fetches the current rates from CRS, inverts them into ET's convention,
// and upserts EUR->RSD / USD->RSD rows. Pairs that CRS did not return are left
// untouched and reported in SyncResult.Missing. A gRPC/transport error is
// returned without touching the database.
func (s *Syncer) Sync(ctx context.Context) (SyncResult, error) {
	now := time.Now().UTC()

	fetched, err := s.fetcher.FetchRates(ctx)
	if err != nil {
		return SyncResult{}, fmt.Errorf("currency sync: fetch: %w", err)
	}

	result := SyncResult{FetchedAt: now}
	seen := make(map[string]bool, len(fetched))

	for _, fr := range fetched {
		// CRS returns base=RSD, so each pair is RSD->foreign. ET stores the
		// inverse: from=foreign, to=RSD, rate = 1 / CRS.rate.
		foreign := fr.To
		seen[foreign] = true

		if fr.Rate == 0 {
			log.Printf("WARN: currency sync: skipping %s->%s, CRS rate is 0 (cannot invert)", fr.From, fr.To)
			continue
		}
		if fr.IsOutdated {
			log.Printf("WARN: currency sync: CRS reports %s->%s rate as outdated (provider=%s), writing anyway", fr.From, fr.To, fr.SourceProvider)
		}

		// Invert with decimal precision, then round to the column scale (6).
		invRate := decimal.NewFromInt(1).Div(decimal.NewFromFloat(fr.Rate)).Round(6)

		// rate.date = the day the rate is valid for, taken from CRS updated_at (UTC).
		rateDate := fr.UpdatedAt.UTC()
		if rateDate.IsZero() {
			rateDate = now
		}

		if _, err := s.repo.Upsert(ctx, repository.CreateRateInput{
			FromCurrency: foreign,
			ToCurrency:   "RSD",
			Rate:         invRate,
			Date:         rateDate,
			Source:       fr.SourceProvider,
			FetchedAt:    now,
		}); err != nil {
			return result, fmt.Errorf("currency sync: upsert %s->RSD: %w", foreign, err)
		}

		invFloat, _ := invRate.Float64()
		result.Pairs = append(result.Pairs, SyncedPair{
			From:   foreign,
			To:     "RSD",
			Rate:   invFloat,
			Source: fr.SourceProvider,
		})
		result.SyncedPairs++
	}

	// Diff requested targets against what CRS returned.
	for _, target := range targetCurrencies {
		if !seen[target] {
			result.Missing = append(result.Missing, target)
		}
	}

	if len(result.Missing) > 0 {
		log.Printf("WARN: currency sync: targets not returned by CRS: %v", result.Missing)
	}
	log.Printf("currency sync: upserted %d pair(s), missing %d, fetched_at=%s",
		result.SyncedPairs, len(result.Missing), now.Format(time.RFC3339))

	return result, nil
}
