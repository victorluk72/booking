package main

import (
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/victorluk72/booking/internal/helpers"
)

// WriteToConcole is a "middleware" type of function that
// will write something to console when page is hit
func WriteToConcole(next http.Handler) http.Handler {

	//Return anonimouse function, cast it to type "http.HandlerFunc"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hit the page")
		next.ServeHTTP(w, r)
	})
}

// NoSurf generates CSRF protection to all POST request
// It is custom built piece of middleware
func NoSurf(next http.Handler) http.Handler {

	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction, //set to true for https
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad loads and saves the session on every request
// With session avaialble we can do many stuff (e.g User auth, pass data between pages etc)
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

//Auth is a middleware function used to protect routes that accesable only to authorized users
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//negative path
		if !helpers.IsAuthenticated(r) {
			session.Put(r.Context(), "error", "Log in first")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
