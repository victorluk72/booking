package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
	"github.com/victorluk72/booking/internal/config"
	"github.com/victorluk72/booking/internal/models"
	"github.com/victorluk72/booking/internal/render"
)

//Variable that controls the session (from package scs)
var session *scs.SessionManager

// Get all configuration values (from package "config")
// Now this variable availabe for whole "main" package
var app config.AppConfig

// This is variable to store patch to templates
// It can be differewnt for Linux
var pathToTemplates = "../../templates"

// Define var "functions". We will use it to allow our own functions in tempalte
// These will be a custom functions for tempaltes (in future)
var functions = template.FuncMap{}

func getRoutes() http.Handler {

	//----Pasted from func run() from main package
	gob.Register(models.Reservation{})

	//Change these to "true" when in Production
	app.InProduction = false

	//Define new INFO and ERROR logger and make it avaialble for whole application (vial app.Infolog)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	//----Session managment-----------------------
	session = scs.New()
	session.Lifetime = 24 * time.Hour              //make session valid for 24 hours
	session.Cookie.Persist = true                  //Pesist session data in the cookies
	session.Cookie.SameSite = http.SameSiteLaxMode //WTF?
	session.Cookie.Secure = app.InProduction       //This is haandling https. Set true for Prod

	// Now asign whatever you have for session in main to app.Config variable
	app.Session = session

	//----Session managment Ends------------------

	//----Template cache managment-------------------
	//Call my template cache (tc) from package render
	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("Can't create template cache", err)
		//return err
	}

	// Assign my tempalte cache to configuration variable app.TemplateCache
	// This allows get cache once and do not reach it every time we browse
	//app.TemplateCache = tc

	//Don't use template cache (for example during dev process)
	// false for Dev, true for Prod and for Test
	app.UseCache = true

	//--TEMP:Print list of all pages from tempalte cache
	fmt.Println("---This is my template cache:---")
	for pg := range tc {
		fmt.Println(pg)
	}
	fmt.Println("---End my template cache:---")
	//--TEMP ENDS:Print list of all pages from tempalte cache

	app.UseCache = false

	//This give render package access to our app variable
	render.NewRenderer(&app)
	//----Tempalte cache managment Ends----------------

	// This is to create repository variable
	repo := NewRipo(&app)
	//Pass it back to handlers (Why?)
	NewHandlers(repo)

	//----Pasted from routs.go file
	mux := chi.NewRouter()

	//--------Middleware block--------------
	//This is middleware to recover from panic
	mux.Use(middleware.Recoverer)

	// This is custom middleware function (see package middleware.go)
	// It means ignore any request if it does't have proper CSRFToken protection
	// If you have form without CSRF protection it will return "Bad request"
	//mux.Use(NoSurf)

	// This is middleware for session load
	mux.Use(SessionLoad)
	//--------Middleware block Ends------------

	//------These are my routes---------------
	mux.Get("/", Ripo.Home)
	mux.Get("/about", Ripo.About)
	mux.Get("/generals", Ripo.Generals)
	mux.Get("/majors", Ripo.Majors)
	mux.Get("/contact", Ripo.Contact)

	mux.Get("/search-availability", Ripo.Availability)
	mux.Post("/search-availability", Ripo.PostAvailability)
	mux.Post("/search-availability-json", Ripo.AvailabilityJSON)

	mux.Get("/make-reservation", Ripo.Reservation)
	mux.Post("/make-reservation", Ripo.PostReservation)
	mux.Get("/reservation-summary", Ripo.ReservationSummary)

	//------End of my routes block---------------

	//Create file server to manage our static files
	fileServer := http.FileServer(http.Dir("../../static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
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
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// CreateTestTemplateCache is to define map that would hold all template files
// from "template" directory index would be a file name and value is pointer to rendered tempalte
// INPORTANT: Using this map with cache prevent from reading templates from disk ever time when page loaded
// instead we read if from "in memory" cache - increase speed dramatically!
func CreateTestTemplateCache() (map[string]*template.Template, error) {

	// This is to hold all tempaltes as a map - pointing to template addresses
	// Index in this map is tempalte name and value is pointer to template
	myCache := map[string]*template.Template{}

	//This give you list of all full path to files that have "page.html" in the file name
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.html", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		//this is extact file name only from pages
		name := filepath.Base(page)

		//This is a tempalte set (ts)  (all tempaltes)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		//Check for base.layout.html files in templates folder
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.html", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.html", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		//Add template set to the myCache variable
		myCache[name] = ts
	}

	return myCache, nil
}
