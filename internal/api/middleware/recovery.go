package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/DigitLock/expense-tracker/internal/dto"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log stack trace
				log.Printf("PANIC: %v\n%s", err, debug.Stack())

				// Return 500 error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				response := dto.NewErrorResponse(
					"INTERNAL_ERROR",
					"An unexpected error occurred",
					nil,
				)
				if err := json.NewEncoder(w).Encode(response); err != nil {
					log.Printf("ERROR encoding response: %v", err)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
