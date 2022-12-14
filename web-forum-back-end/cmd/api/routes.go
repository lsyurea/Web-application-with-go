package main

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	// define routes
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Get("/", app.Home)

	mux.Get("/posts", app.GetPosts)
	return mux
}