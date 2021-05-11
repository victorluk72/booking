package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/victorluk72/booking/internal/config"
	"github.com/victorluk72/booking/internal/handlers"
)

// routes ... returns http Handler
func routes(app *config.AppConfig) http.Handler {

	mux := chi.NewRouter()

	//--------Middleware block--------------
	//This is middleware to recover from panic
	mux.Use(middleware.Recoverer)

	// This is custom middleware function (see package middleware.go)
	// It means ignore any request if it does't have proper CSRFToken protection
	// If you have form without CSRF protection it will return "Bad request"
	mux.Use(NoSurf)

	// This is middleware for session load
	mux.Use(SessionLoad)
	//--------Middleware block End------------

	//------These are my routes---------------
	mux.Get("/", handlers.Ripo.Home)
	mux.Get("/about", handlers.Ripo.About)
	mux.Get("/generals", handlers.Ripo.Generals)
	mux.Get("/majors", handlers.Ripo.Majors)
	mux.Get("/contact", handlers.Ripo.Contact)

	mux.Get("/search-availability", handlers.Ripo.Availability)
	mux.Post("/search-availability", handlers.Ripo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Ripo.AvailabilityJSON)

	mux.Get("/make-reservation", handlers.Ripo.Reservation)
	//------End of my routes block---------------

	//Create file server to manage our static files
	fileServer := http.FileServer(http.Dir("../../static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux

}
