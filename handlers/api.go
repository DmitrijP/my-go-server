package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/DmitrijP/my-go-server/internal/auth"
	"github.com/DmitrijP/my-go-server/internal/database"
)

type user_create struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type user_create_response struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
}

func ReadinessHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *ApiConfig) MetricsShow(w http.ResponseWriter, req *http.Request) {
	val := cfg.FileserverHits.Load()
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

func (cfg *ApiConfig) UsersHandler(w http.ResponseWriter, req *http.Request) {

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
		return
	}

	user, err := cfg.Db.CreateUser(req.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hpass})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondWithError(w, http.StatusConflict, "User may already exist")
		return
	}
	resObj := user_create_response{Id: user.ID.String(), CreatedAt: user.CreatedAt.String(), UpdatedAt: user.UpdatedAt.String(), Email: user.Email}
	respondWithJSON(w, http.StatusCreated, resObj)
}
