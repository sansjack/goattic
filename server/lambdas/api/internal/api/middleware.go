package api

import (
	"context"
	"net/http"

	"goattic-api/internal/auth"
	"goattic-api/internal/storage"
)

type contextKey string

const apiKeyContextKey contextKey = "apiKey"

func AuthMiddleware(db *storage.DynamoDBClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			apiKey, err := auth.ExtractAPIKey(r.Header.Get("X-Api-Key"))

			if err != nil {
				respondError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
				return
			}

			keyData, err := db.ValidateAPIKey(r.Context(), apiKey)
			if err != nil {
				respondError(w, http.StatusUnauthorized, "Unauthorized: invalid API key")
				return
			}

			if !keyData.Enabled {
				respondError(w, http.StatusUnauthorized, "Unauthorized: API key is disabled")
				return
			}

			ctx := context.WithValue(r.Context(), apiKeyContextKey, keyData)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAPIKeyFromContext(ctx context.Context) *storage.APIKey {
	if key, ok := ctx.Value(apiKeyContextKey).(*storage.APIKey); ok {
		return key
	}
	return nil
}
