package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"goattic-api/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sixafter/nanoid"
)

type API struct {
	db          *storage.DynamoDBClient
	s3          *storage.S3Client
	mediaDomain string
}

func NewAPI(db *storage.DynamoDBClient, s3 *storage.S3Client, mediaDomain string) *API {
	return &API{
		db:          db,
		s3:          s3,
		mediaDomain: mediaDomain,
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

		r.Get("/key", a.handleGetKey)
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

func (a *API) handleGetKey(w http.ResponseWriter, r *http.Request) {
	apiKey := GetAPIKeyFromContext(r.Context())
	if apiKey == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"id":        apiKey.ID,
		"owner":     apiKey.Owner,
		"createdAt": apiKey.CreatedAt,
		"rateLimit": apiKey.RateLimit,
		"enabled":   apiKey.Enabled,
	})
}

func (a *API) handleUpload(w http.ResponseWriter, r *http.Request) {
	apiKey := GetAPIKeyFromContext(r.Context())
	if apiKey == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Filename string `json:"filename"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Filename == "" {
		respondError(w, http.StatusBadRequest, "filename is required")
		return
	}

	ext := filepath.Ext(req.Filename)
	id, err := nanoid.NewWithLength(12)

	if err != nil {
		panic(err)
	}
	key := fmt.Sprintf("%s/%s%s", apiKey.ID, id.String(), ext)

	presignedPost, err := a.s3.GeneratePresignedPostURL(r.Context(), key)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate upload URL")
		return
	}

	publicURL := fmt.Sprintf("https://%s/%s", a.mediaDomain, key)

	respondJSON(w, http.StatusOK, map[string]any{
		"presignedPost": map[string]any{
			"url":    presignedPost.URL,
			"fields": presignedPost.Values,
		},
		"key":       key,
		"publicUrl": publicURL,
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
