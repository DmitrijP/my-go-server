package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/DmitrijP/my-go-server/internal/auth"
	"github.com/DmitrijP/my-go-server/internal/database"
	"github.com/google/uuid"
)

type polka_event struct {
	Event string    `json:"event"`
	Data  user_data `json:"data"`
}

type user_data struct {
	UserId string `json:"user_id"`
}

func (cfg *ApiConfig) PolkaWebhookHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	key, err := auth.GetAPIKey(req.Header)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Something went wrong")
		return
	}

	if key != cfg.PolkaKey {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Wrong API KEY")
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := polka_event{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusNoContent, "Something went wrong")
		return
	}

	if params.Event != "user.upgraded" {
		respondWithoutBody(w, http.StatusNoContent)
		return
	}
	uid, err := uuid.Parse(params.Data.UserId)
	if err != nil {
		respondWithoutBody(w, http.StatusNoContent)
		return
	}
	updtUsr, err := cfg.Db.SelectUserById(req.Context(), uid)
	if err != nil {
		log.Printf("Error selecting usr: %s", err)
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	p := database.UpdateChirpyRedParams{ID: updtUsr.ID, IsChirpyRed: true}
	err = cfg.Db.UpdateChirpyRed(req.Context(), p)
	if err != nil {
		log.Printf("Error selecting usr: %s", err)
		respondWithError(w, http.StatusNotFound, err.Error())
	}

	respondWithoutBody(w, http.StatusNoContent)
}
