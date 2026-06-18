package currency

import (
	"context"
	"log"
	"time"
)

// Scheduler runs periodic rate syncs in a single background goroutine.
type Scheduler struct {
	syncer   *Syncer
	interval time.Duration
	done     chan struct{} // closed when the goroutine exits
}

// NewScheduler wires a Syncer with the sync interval.
func NewScheduler(syncer *Syncer, interval time.Duration) *Scheduler {
	return &Scheduler{syncer: syncer, interval: interval, done: make(chan struct{})}
}

// Start launches the background sync loop in its own goroutine and returns
// immediately (does not block the caller / server startup). It performs an
// immediate first sync, then one sync per interval until ctx is cancelled.
//
// A sync error is only logged as a warning — it never crashes the process, so
// the loop survives CRS being unavailable. The ticker does not fire at t=0; the
// initial sync covers that, avoiding a duplicate sync on startup.
func (s *Scheduler) Start(ctx context.Context) {
	go func() {
		defer close(s.done)

		if _, err := s.syncer.Sync(ctx); err != nil {
			log.Printf("WARN: initial currency sync failed: %v", err)
		}

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if _, err := s.syncer.Sync(ctx); err != nil {
					log.Printf("WARN: currency sync failed: %v", err)
				}
			case <-ctx.Done():
				log.Println("currency scheduler stopped")
				return
			}
		}
	}()
}

// Wait blocks until the scheduler goroutine has fully exited. Call it after
// cancelling the context passed to Start, before closing the client, to avoid
// racing Close against an in-flight sync.
func (s *Scheduler) Wait() {
	<-s.done
}
