package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/DmitrijP/my-go-server/internal/auth"
	"github.com/DmitrijP/my-go-server/internal/database"
	"github.com/google/uuid"
)

type chirp_model struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body      string `json:"body"`
	UserId    string `json:"user_id"`
}

type chirp_create struct {
	Body string `json:"body"`
}

func (cfg *ApiConfig) ChirpsHandler(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error fetching Bearer Token: %s", err)
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	user_id, err := auth.ValidateJWT(token, cfg.Jwt_secret)
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

	var c = database.CreateChirpParams{Body: lowerBody, UserID: user_id}

	chirp, err := cfg.Db.CreateChirp(req.Context(), c)
	if err != nil {
		log.Printf("Error saving chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	resObj := chirp_model{Id: chirp.ID.String(), CreatedAt: chirp.CreatedAt.String(), UpdatedAt: chirp.UpdatedAt.String(), Body: chirp.Body, UserId: chirp.UserID.String()}
	respondWithJSON(w, http.StatusCreated, resObj)
}

func (cfg *ApiConfig) GetAllChirpsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	s := req.URL.Query().Get("author_id")
	qId, err := uuid.Parse(s)
	var chirps []database.Chirp
	if s != "" && err != nil {
		chirps, _ = cfg.Db.GetAllChirpsByAuthor(req.Context(), qId)
	} else {
		chirps, _ = cfg.Db.GetAllChirps(req.Context())
	}

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

func (cfg *ApiConfig) GetOneChirpsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chirpId := req.PathValue("chirpID")
	chirpUuid, err := uuid.Parse(chirpId)
	if err != nil {
		log.Printf("Error parsing chirp id: %s", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirp, err := cfg.Db.GetOneChirp(req.Context(), chirpUuid)
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

func (cfg *ApiConfig) DeleteChirpHandler(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error fetching Bearer Token: %s", err)
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	id, err := auth.ValidateJWT(token, cfg.Jwt_secret)
	if err != nil {
		log.Printf("Error validating jwt: %s", err)
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	chirpId := req.PathValue("chirpID")
	chirpUuid, err := uuid.Parse(chirpId)
	if err != nil {
		log.Printf("Error parsing chirp id: %s", err)
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	chirp, err := cfg.Db.GetOneChirp(req.Context(), chirpUuid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	if id != chirp.UserID {
		respondWithError(w, http.StatusForbidden, "Forbidden to delete foreign chirps")
		return
	}

	err = cfg.Db.DeleteChirp(req.Context(), chirpUuid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithoutBody(w, http.StatusNoContent)
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
