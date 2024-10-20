package handlers

import (
	"net/http"
	"os"
)

func (cfg *ApiConfig) MetricsReset(w http.ResponseWriter, req *http.Request) {
	p := os.Getenv("PLATFORM")
	if p != "dev" {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		return
	}
	cfg.FileserverHits.Swap(0)
	err := cfg.Db.DeleteAllUsers(req.Context())
	if err != nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
