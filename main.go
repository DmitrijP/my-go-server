package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/DmitrijP/my-go-server/handlers"
	"github.com/DmitrijP/my-go-server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type parameters struct {
	Body string `json:"body"`
}

func main() {
	godotenv.Load()
	jwt_secret := os.Getenv("JWT_SECRET")

	dbURL := os.Getenv("DB_URL")
	db, _ := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	var cfg handlers.ApiConfig
	cfg.Db = *dbQueries
	cfg.Jwt_secret = jwt_secret

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", handlers.ReadinessHandler)

	mux.HandleFunc("POST /api/login", cfg.LoginHandler)
	mux.HandleFunc("POST /api/refresh", cfg.RefreshHandler)
	mux.HandleFunc("POST /api/revoke", cfg.RevokeHandler)

	mux.HandleFunc("POST /api/users", cfg.UsersHandler)
	mux.HandleFunc("PUT /api/users", cfg.ChangeUserPasswordHandler)

	mux.HandleFunc("POST /api/chirps", cfg.ChirpsHandler)
	mux.HandleFunc("GET /api/chirps", cfg.GetAllChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.GetOneChirpsHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.DeleteChirpHandler)

	mux.HandleFunc("POST /admin/reset", cfg.MetricsReset)
	mux.HandleFunc("GET /admin/metrics", cfg.MetricsShow)

	mux.Handle("/app/", cfg.MiddlewareMetricsInc(
		http.StripPrefix("/app/",
			http.FileServer(http.Dir("./html/")))))

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
