package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/victorluk72/booking/pkg/config"
	"github.com/victorluk72/booking/pkg/handlers"
)

// routes ... returns http Handler
func routes(app *config.AppConfig) http.Handler {

	mux := chi.NewRouter()

	//--------Middleware block--------------
	//This is middleware to recover from panic
	mux.Use(middleware.Recoverer)

	// This is custom middleware function (see package middleware.go)
	mux.Use(NoSurf)

	// This is middleware for session load
	mux.Use(SessionLoad)
	//--------Middleware block End------------

	//These are my routes
	mux.Get("/", handlers.Ripo.Home)
	mux.Get("/about", handlers.Ripo.About)

	//Create file server to manage our static files
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux

}
