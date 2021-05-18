package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/victorluk72/booking/internal/config"
	"github.com/victorluk72/booking/internal/handlers"
	"github.com/victorluk72/booking/internal/models"
	"github.com/victorluk72/booking/internal/render"
)

const portNumber = ":8080"

//Varuiable that controls the session (from package scs)
var session *scs.SessionManager

// Get all configuration values (from package "config")
// Now this variable availabe for whole "main" package
var app config.AppConfig

func main() {

	// What I'm going to put into my session? - I'm passing Reservation model
	// You need this to pass the content of the reservation form to page reservation-summary
	// See handlers.go m.App.Session.Put(r.Context(), "reservation-details", reservation)
	gob.Register(models.Reservation{})

	//Change these to "true" when in Production
	app.InProduction = false

	//----Session managment-----------------------
	session = scs.New()
	session.Lifetime = 24 * time.Hour              //make session valid for 24 hours
	session.Cookie.Persist = true                  //Pesist session data in the cookies
	session.Cookie.SameSite = http.SameSiteLaxMode //WTF?
	session.Cookie.Secure = app.InProduction       //This is haandling https. Set true for Prod

	// Now asign whatever you have for session in main to app.Config variable
	app.Session = session

	//----Session managment Ends------------------

	//----Tempalte cache managment-------------------
	//Call my template cache (tc) from package render
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Can't create template cache", err)
	}

	// Assign my tempalte cache to configuration variable app.TemplateCache
	// This allows get cache once and do not reach it every time we browse
	app.TemplateCache = tc

	//Don't use template cache (for example during dev process)
	// false for Dev, true for Prod
	app.UseCache = false

	//TEMP:Print list of all pages from tempalte cache
	fmt.Println("---This is my template cache:---")
	for pg := range tc {
		fmt.Println(pg)
	}
	fmt.Println("---End my template cache:---")

	app.UseCache = false

	//This give render package access to our app variable
	render.NewTemplates(&app)
	//----Tempalte cache managment Ends----------------

	// This is to create repository variable
	repo := handlers.NewRipo(&app)
	//Pass it back to handlers (Why?)
	handlers.NewHandlers(repo)

	fmt.Println("...Starting applicaton on port", portNumber)

	// Define my http Server
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	//Run web server that would listen and serve
	err = srv.ListenAndServe()
	log.Fatal(err)

}
