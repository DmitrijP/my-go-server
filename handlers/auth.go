package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/DmitrijP/my-go-server/internal/auth"
	"github.com/DmitrijP/my-go-server/internal/database"
)

type refresh_model struct {
	RefreshToken string `json:"refresh_token"`
}

type auth_model struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type user_model struct {
	Id           string `json:"id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (cfg *ApiConfig) LoginHandler(w http.ResponseWriter, req *http.Request) {

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

	usr, err := cfg.Db.SelectUserByEmail(req.Context(), params.Email)
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

	token, err := auth.MakeJWT(usr.ID, cfg.Jwt_secret, time.Duration(expirationTime)*time.Second)
	if err != nil {
		log.Printf("Error creating jwt: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	refresh, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error creating jwt: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	tkParams := database.CreateTokenParams{
		UserID:    usr.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	_, err = cfg.Db.CreateToken(req.Context(), tkParams)
	if err != nil {
		log.Printf("Error creating refresh: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	resObj := user_model{
		Id:           usr.ID.String(),
		CreatedAt:    usr.CreatedAt.String(),
		UpdatedAt:    usr.UpdatedAt.String(),
		Email:        usr.Email,
		Token:        token,
		RefreshToken: refresh,
	}
	respondWithJSON(w, http.StatusOK, resObj)
}

func (cfg *ApiConfig) RefreshHandler(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error fetching Bearer Token: %s", err)
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	tok, err := cfg.Db.GetOneToken(req.Context(), token)
	if err != nil {
		log.Printf("Error selecting token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}

	if tok.ExpiresAt.Before(time.Now()) {
		log.Printf("Refresh expired: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}

	if tok.RevokedAt.Valid && tok.RevokedAt.Time.After(time.Now()) {
		log.Printf("Refresh revoked: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Refresh token revoked")
		return
	}

	err = cfg.Db.RevokeToken(req.Context(), tok.Token)
	if err != nil {
		log.Printf("Refresh revokation failed: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Something went wrong")
		return
	}

	jwt, err := auth.MakeJWT(tok.UserID, cfg.Jwt_secret, time.Hour)
	if err != nil {
		log.Printf("New JWT creation failed: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Something went wrong")
		return
	}

	jwts := struct {
		Token string `json:"token"`
	}{Token: jwt}

	respondWithJSON(w, http.StatusOK, jwts)
}

func (cfg *ApiConfig) RevokeHandler(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error fetching Bearer Token: %s", err)
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	tok, err := cfg.Db.GetOneToken(req.Context(), token)
	if err != nil {
		log.Printf("Error selecting token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}

	if tok.ExpiresAt.Before(time.Now()) {
		log.Printf("Refresh expired: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}

	err = cfg.Db.RevokeToken(req.Context(), tok.Token)
	if err != nil {
		log.Printf("Refresh revokation failed: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}
	respondWithoutBody(w, http.StatusNoContent)
}
