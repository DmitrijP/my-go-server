package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/DmitrijP/my-go-server/internal/auth"
	"github.com/DmitrijP/my-go-server/internal/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	jwt_secret     string
	fileserverHits atomic.Int32
	db             database.Queries
}

type parameters struct {
	Body string `json:"body"`
}

type user_model struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
	Token     string `json:"token"`
}

type chirp_model struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body      string `json:"body"`
	UserId    string `json:"user_id"`
}

type chirp_create struct {
	Body   string `json:"body"`
	UserId string `json:"user_id"`
}

type auth_model struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ExpireInSeconds *int   `json:"expires_in_seconds,omitempty"`
}

type user_create struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type http_error struct {
	Error string `json:"error"`
}

type http_resp struct {
	CleanedBody string `json:"cleaned_body"`
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
	p := os.Getenv("PLATFORM")
	if p != "dev" {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		return
	}
	cfg.fileserverHits.Swap(0)
	err := cfg.db.DeleteAllUsers(req.Context())
	if err != nil {
		//TODO handle
	}
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
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

var forbiddenWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

func cleanChirpText(input string) string {
	words := strings.Split(input, " ")
	for i, word := range words {
		for _, forbidden := range forbiddenWords {
			word = strings.ToLower(word)
			if word == forbidden {
				words[i] = "****"
			}
		}
	}
	input = strings.Join(words, " ")
	return input
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	params := auth_model{}
	err := decoder.Decode(&params)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	expirationTime := 60 * 60
	if params.ExpireInSeconds != nil && *params.ExpireInSeconds < expirationTime {
		expirationTime = *params.ExpireInSeconds
	}

	usr, err := cfg.db.SelectUserByEmail(req.Context(), params.Email)
	if err != nil {
		log.Printf("Error selecting user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	err = auth.CheckPasswordHash(params.Password, usr.HashedPassword)
	if err != nil {
		log.Printf("Error comparing passwords: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	token, err := auth.MakeJWT(usr.ID, cfg.jwt_secret, time.Duration(expirationTime)*time.Second)
	if err != nil {
		log.Printf("Error creating jwt: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	resObj := user_model{
		Id:        usr.ID.String(),
		CreatedAt: usr.CreatedAt.String(),
		UpdatedAt: usr.UpdatedAt.String(),
		Email:     usr.Email,
		Token:     token,
	}
	respondWithJSON(w, http.StatusOK, resObj)
}

func (cfg *apiConfig) usersHandler(w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	params := user_create{}
	err := decoder.Decode(&params)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	hpass, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Unknown Error")
	}

	user, _ := cfg.db.CreateUser(req.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hpass})

	resObj := user_model{Id: user.ID.String(), CreatedAt: user.CreatedAt.String(), UpdatedAt: user.UpdatedAt.String(), Email: user.Email}
	respondWithJSON(w, http.StatusCreated, resObj)
}

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error fetching Bearer Token: %s", err)
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	_, err = auth.ValidateJWT(token, cfg.jwt_secret)
	if err != nil {
		log.Printf("Error validating jwt: %s", err)
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	decoder := json.NewDecoder(req.Body)
	params := chirp_create{}
	err = decoder.Decode(&params)
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

	lowerBody := cleanChirpText(params.Body)

	u, err := uuid.Parse(params.UserId)
	var c = database.CreateChirpParams{Body: lowerBody, UserID: u}

	chirp, err := cfg.db.CreateChirp(req.Context(), c)
	if err != nil {
		log.Printf("Error saving chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	resObj := chirp_model{Id: chirp.ID.String(), CreatedAt: chirp.CreatedAt.String(), UpdatedAt: chirp.UpdatedAt.String(), Body: chirp.Body, UserId: chirp.UserID.String()}
	respondWithJSON(w, http.StatusCreated, resObj)
}

func (cfg *apiConfig) getAllChirpsHandler(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	chirps, _ := cfg.db.GetAllChirps(req.Context())
	var chirp_models []chirp_model
	for _, chirp := range chirps {
		chirp_models = append(chirp_models,
			chirp_model{
				Id:        chirp.ID.String(),
				CreatedAt: chirp.CreatedAt.String(),
				UpdatedAt: chirp.UpdatedAt.String(),
				Body:      chirp.Body,
				UserId:    chirp.UserID.String(),
			})
	}

	respondWithJSON(w, http.StatusOK, chirp_models)
}

func (cfg *apiConfig) getOneChirpsHandler(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	chirpId := req.PathValue("chirpID")
	chirpUuid, err := uuid.Parse(chirpId)
	if err != nil {
		//TODO
	}

	chirp, err := cfg.db.GetOneChirp(req.Context(), chirpUuid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	chirp_model := chirp_model{
		Id:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.String(),
		UpdatedAt: chirp.UpdatedAt.String(),
		Body:      chirp.Body,
		UserId:    chirp.UserID.String(),
	}
	respondWithJSON(w, http.StatusOK, chirp_model)
}

func main() {
	godotenv.Load()
	jwt_secret := os.Getenv("JWT_SECRET")

	dbURL := os.Getenv("DB_URL")
	db, _ := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	var cfg apiConfig
	cfg.db = *dbQueries
	cfg.jwt_secret = jwt_secret

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", readinessHandler)

	mux.HandleFunc("POST /api/login", cfg.loginHandler)

	mux.HandleFunc("POST /api/users", cfg.usersHandler)

	mux.HandleFunc("POST /api/chirps", cfg.chirpsHandler)

	mux.HandleFunc("GET /api/chirps", cfg.getAllChirpsHandler)

	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getOneChirpsHandler)

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
