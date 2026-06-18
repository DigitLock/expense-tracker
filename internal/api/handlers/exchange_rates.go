package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/DigitLock/expense-tracker/internal/currency"
	"github.com/DigitLock/expense-tracker/internal/dto"
)

// ExchangeRateHandler exposes operations on exchange rates. It shares the same
// Syncer as the background scheduler. syncer may be nil when the currency
// client failed to initialize at startup.
type ExchangeRateHandler struct {
	syncer *currency.Syncer
}

func NewExchangeRateHandler(syncer *currency.Syncer) *ExchangeRateHandler {
	return &ExchangeRateHandler{syncer: syncer}
}

// Sync godoc
// @Summary      Force exchange rate sync
// @Description  Triggers an immediate sync of exchange rates from the currency rate service and upserts them.
// @Tags         Currencies
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.SuccessResponse{data=dto.ExchangeRateSyncResponse} "Sync result"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      503 {object} dto.ErrorResponse "Currency rate service unavailable"
// @Router       /exchange-rates/sync [post]
func (h *ExchangeRateHandler) Sync(w http.ResponseWriter, r *http.Request) {
	if h.syncer == nil {
		writeError(w, http.StatusServiceUnavailable, "CURRENCY_SERVICE_UNAVAILABLE",
			"Currency rate service is not configured")
		return
	}

	result, err := h.syncer.Sync(r.Context())
	if err != nil {
		// Fetch failed (e.g. CRS down). Sync does not write on fetch error, so
		// the last known rates are preserved.
		writeError(w, http.StatusServiceUnavailable, "CURRENCY_SERVICE_UNAVAILABLE",
			"Failed to sync exchange rates from currency service")
		return
	}

	writeSuccess(w, http.StatusOK, dto.ExchangeRateSyncResponse{
		SyncedPairs: result.SyncedPairs,
		Source:      distinctSources(result.Pairs),
		FetchedAt:   result.FetchedAt,
	})
}

// distinctSources returns the single provider name if all synced pairs share
// one, otherwise the distinct providers joined by ", ".
func distinctSources(pairs []currency.SyncedPair) string {
	set := make(map[string]struct{}, len(pairs))
	for _, p := range pairs {
		if p.Source != "" {
			set[p.Source] = struct{}{}
		}
	}
	sources := make([]string, 0, len(set))
	for s := range set {
		sources = append(sources, s)
	}
	sort.Strings(sources)
	return strings.Join(sources, ", ")
}
