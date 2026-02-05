package api

import (
	"encoding/json"
	"net/http"

	"goattic-lambda/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	db *storage.DynamoDBClient
	s3 *storage.S3Client
}

func NewAPI(db *storage.DynamoDBClient, s3 *storage.S3Client) *API {
	return &API{
		db: db,
		s3: s3,
	}
}

func (a *API) Router() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", a.handleRoot)
	r.Get("/health", a.handleHealth)

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(a.db))

		r.Get("/keys", a.handleListKeys)
		r.Post("/upload", a.handleUpload)
	})

	return r
}

func (a *API) handleRoot(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"version": "1.0.0",
	})
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

func (a *API) handleListKeys(w http.ResponseWriter, r *http.Request) {
	apiKey := GetAPIKeyFromContext(r.Context())
	if apiKey == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	keys, err := a.db.ListAPIKeysByOwner(r.Context(), apiKey.Owner)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list keys")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"keys": keys,
	})
}

func (a *API) handleUpload(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Upload endpoint - implement file upload logic",
	})
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}

func respondJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
