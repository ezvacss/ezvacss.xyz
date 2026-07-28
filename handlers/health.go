// handlers/health.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

func HealthCheck(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			if encErr := json.NewEncoder(w).Encode(map[string]string{"status": "db unreachable"}); encErr != nil {
				log.Printf("healthz: failed to write response: %v", encErr)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			log.Printf("healthz: failed to write response: %v", err)
		}
	}
}
