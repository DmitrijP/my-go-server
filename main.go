package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

type parameters struct {
	Body string `json:"body"`
}

type http_error struct {
	Error string `json:"error"`
}

type http_resp struct {
	Valid bool `json:"valid"`
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

func respondWithError(w http.ResponseWriter, code int, msg string) {
	errorObj := http_error{Error: msg}
	dat, err := json.Marshal(errorObj)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(code)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func validateChirpHandler(w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	resObj := http_resp{Valid: true}
	respondWithJSON(w, http.StatusOK, resObj)
}

func main() {
	var cfg apiConfig

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", readinessHandler)

	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

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
