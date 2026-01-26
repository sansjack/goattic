package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	bucketName string
	tableName  string
	router     *chi.Mux
}

func NewAPI() *API {
	api := &API{
		bucketName: os.Getenv("BUCKET_NAME"),
		tableName:  os.Getenv("TABLE_NAME"),
		router:     chi.NewRouter(),
	}

	api.router.Use(middleware.Logger)
	api.router.Use(middleware.Recoverer)

	api.setupRoutes()

	return api
}

func (a *API) setupRoutes() {
	a.router.Get("/", a.handleRoot)
	a.router.Get("/health", a.handleHealth)

	a.router.Route("/api", func(r chi.Router) {
		r.Get("/items", a.handleListItems)
		r.Post("/items", a.handleCreateItem)
		r.Get("/items/{id}", a.handleGetItem)
		r.Put("/items/{id}", a.handleUpdateItem)
		r.Delete("/items/{id}", a.handleDeleteItem)
	})
}

func (a *API) handleRoot(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "GoAttic API",
		"version": "1.0.0",
	})
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{
		"status":     "healthy",
		"bucketName": a.bucketName,
		"tableName":  a.tableName,
	})
}

func (a *API) handleListItems(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "List items",
	})
}

func (a *API) handleCreateItem(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusCreated, map[string]string{
		"message": "Item created",
	})
}

func (a *API) handleGetItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Get item",
		"id":      id,
	})
}

func (a *API) handleUpdateItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Item updated",
		"id":      id,
	})
}

func (a *API) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Item deleted",
		"id":      id,
	})
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func main() {
	api := NewAPI()
	lambda.Start(httpadapter.NewV2(api.router).ProxyWithContext)
}
