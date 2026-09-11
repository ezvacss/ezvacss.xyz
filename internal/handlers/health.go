package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"ezvacss.xyz/internal/db"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if db.Pool == nil || db.Pool.Ping(r.Context()) != nil {
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
