package observability

import (
	"encoding/json"
	"net/http"
)

// HealthCheckHandler backs /healthz — used by Docker healthchecks and
// Kubernetes readiness/liveness probes to confirm the process is alive.
// This only reports process liveness, not deep dependency health (Postgres,
// Kafka) — keeping it cheap so probes don't add load or false-negative on
// transient downstream blips.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "payment",
	})
}