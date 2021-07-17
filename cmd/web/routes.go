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
	//--------Middleware block Ends------------

	//------These are my routes---------------
	mux.Get("/", handlers.Ripo.Home)
	mux.Get("/about", handlers.Ripo.About)
	mux.Get("/generals", handlers.Ripo.Generals)
	mux.Get("/majors", handlers.Ripo.Majors)
	mux.Get("/contact", handlers.Ripo.Contact)

	mux.Get("/user/login", handlers.Ripo.Login)
	mux.Post("/user/login", handlers.Ripo.PostLogin)
	mux.Get("/user/logout", handlers.Ripo.Logout)

	mux.Get("/search-availability", handlers.Ripo.Availability)
	mux.Post("/search-availability", handlers.Ripo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Ripo.AvailabilityJSON)
	mux.Get("/choose-room/{id}", handlers.Ripo.ChooseRoom)
	mux.Get("/book-room", handlers.Ripo.BookRoom)

	mux.Get("/make-reservation", handlers.Ripo.Reservation)
	mux.Post("/make-reservation", handlers.Ripo.PostReservation)
	mux.Get("/reservation-summary", handlers.Ripo.ReservationSummary)

	//This is protected area - only for Auth users
	// The "admin" wil lbe cerated automatically to the route
	mux.Route("/admin", func(mux chi.Router) {
		//mux.Use(Auth)

		//This is my protected route
		mux.Get("/dashboard", handlers.Ripo.AdminDashboard)
		mux.Get("/reservations-new", handlers.Ripo.AdminNewReservations)
		mux.Get("/reservations-all", handlers.Ripo.AdminAllReservations)
		mux.Get("/reservations/{src}/{id}", handlers.Ripo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", handlers.Ripo.AdminPostShowReservation)
		mux.Get("/process-reservation/{src}/{id}", handlers.Ripo.AdminProcessReservation)
		mux.Get("/delete-reservation/{src}/{id}", handlers.Ripo.AdminDeleteReservation)

		mux.Get("/reservation-calendar", handlers.Ripo.AdminCalendar)

	})

	//------End of my routes block---------------

	//Create file server to manage our static files
	fileServer := http.FileServer(http.Dir("../../static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux

}
