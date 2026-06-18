// Package currency is ET's gRPC client for the Currency Rate Service (CRS).
//
// In st2 the client only fetches rates as-is from CRS into a domain type. No
// inversion or mapping into ET's exchange_rates convention happens here — that
// is st3. Scheduling/triggering is st4/st5.
package currency

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/DigitLock/expense-tracker/internal/grpc/pb"
)

// FetchedRate is the domain representation of a single rate from CRS, decoupled
// from the generated protobuf type. Values are exactly what CRS returned.
type FetchedRate struct {
	From           string
	To             string
	Rate           float64
	UpdatedAt      time.Time
	SourceProvider string
	IsOutdated     bool
}

// Client is a reusable gRPC client for the Currency Rate Service. It owns a
// single long-lived connection; call Close when done.
type Client struct {
	conn    *grpc.ClientConn
	rpc     pb.CurrencyRateServiceClient
	timeout time.Duration
}

// NewClient dials the currency service at addr (plaintext, no auth) and returns
// a reusable client. The connection is lazy; dialing errors surface on first RPC.
func NewClient(addr string, timeout time.Duration) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("currency: dial %q: %w", addr, err)
	}

	return &Client{
		conn:    conn,
		rpc:     pb.NewCurrencyRateServiceClient(conn),
		timeout: timeout,
	}, nil
}

// Close releases the underlying gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// FetchRates fetches the current RSD->EUR and RSD->USD rates from CRS in a
// single GetRates call and maps them to FetchedRate as-is (no inversion). The
// per-call timeout from the config bounds the request. gRPC errors are returned
// to the caller; there is no retry (that is st4).
func (c *Client) FetchRates(ctx context.Context) ([]FetchedRate, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.rpc.GetRates(callCtx, &pb.GetRatesRequest{
		BaseCurrency:     "RSD",
		TargetCurrencies: []string{"EUR", "USD"},
	})
	if err != nil {
		return nil, fmt.Errorf("currency: GetRates: %w", err)
	}

	rates := make([]FetchedRate, 0, len(resp.GetRates()))
	for _, r := range resp.GetRates() {
		var updatedAt time.Time
		if ts := r.GetUpdatedAt(); ts != nil {
			updatedAt = ts.AsTime()
		}
		rates = append(rates, FetchedRate{
			From:           r.GetFromCurrency(),
			To:             r.GetToCurrency(),
			Rate:           r.GetRate(),
			UpdatedAt:      updatedAt,
			SourceProvider: r.GetSourceProvider(),
			IsOutdated:     r.GetIsOutdated(),
		})
	}

	return rates, nil
}
