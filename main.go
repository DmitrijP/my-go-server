package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func (cfg *apiConfig) metricsShow(w http.ResponseWriter, req *http.Request) {
	val := cfg.fileserverHits.Load()
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	stringVal := fmt.Sprintf(`
		<html>
  		<body>
    		<h1>Welcome, Chirpy Admin</h1>
    		<p>Chirpy has been visited %d times!</p>
  		</body>
		</html>
	`, val)
	w.Write([]byte(stringVal))
}

func (cfg *apiConfig) metricsReset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Swap(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func readinessHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	var cfg apiConfig

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", readinessHandler)

	mux.HandleFunc("POST /admin/reset", cfg.metricsReset)

	mux.HandleFunc("GET /admin/metrics", cfg.metricsShow)

	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./html/")))))

	srv := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	fmt.Println("Starting server on :8080")
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Server start error: %v", err)
	}
}
