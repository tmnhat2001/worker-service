package api

import (
	"context"
	"encoding/json"
	"net/http"
)

type contextKey string

func userFromContext(ctx context.Context) (*User, bool) {
	key := contextKey("user")
	user, ok := ctx.Value(key).(*User)

	if !ok {
		return nil, false
	}

	return user, true
}

func errorResponse(w http.ResponseWriter, errorMessage string, statusCode int) {
	payload := map[string]string{"error": errorMessage}
	jsonResponse(w, payload, statusCode)
}

func jsonResponse(w http.ResponseWriter, payload interface{}, statusCode int) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}
