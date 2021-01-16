package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type contextKey string

func userFromContext(ctx context.Context) (*User, error) {
	key := contextKey("user")
	user, ok := ctx.Value(key).(*User)

	if !ok {
		return nil, errors.New("Unable to retrieve user")
	}

	return user, nil
}

func errorResponse(w http.ResponseWriter, errorMessage string, statusCode int) {
	payload := map[string]string{"error": errorMessage}
	jsonResponse(w, payload, statusCode)
}

func jsonResponse(w http.ResponseWriter, payload interface{}, statusCode int) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}
