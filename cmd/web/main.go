package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/victorluk72/booking/internal/config"
	"github.com/victorluk72/booking/internal/driver"
	"github.com/victorluk72/booking/internal/handlers"
	"github.com/victorluk72/booking/internal/helpers"
	"github.com/victorluk72/booking/internal/models"
	"github.com/victorluk72/booking/internal/render"
)

const portNumber = ":8080"

//Varuiable that controls the session (from package scs)
var session *scs.SessionManager

// Variables for info and erro logs
var infoLog *log.Logger
var errorLog *log.Logger

// Get all configuration values (from package "config")
// Now this variable availabe for whole "main" package
var app config.AppConfig

func main() {

	//Entry point of application

	//return db connection and erro from function run()
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}

	//CLose connection to DB (any type)
	defer db.SQL.Close()

	//Close connection to my mailChan (chennal created for sending emails)
	defer close(app.MailChan)

	//Start listenning my email messages (go routine from send email.go)
	fmt.Println("...Starting email listener....")
	listenForMail()

	fmt.Println("...Starting applicaton on port", portNumber, "...")

	// Define my http Server
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	//Run web server that would listen and serve
	err = srv.ListenAndServe()
	log.Fatal(err)

}

func run() (*driver.DB, error) {
	// What I'm going to put into my session? - I'm passing Reservation, User, Room models
	// You need this to pass the content of the reservation form to page reservation-summary
	// See handlers.go m.App.Session.Put(r.Context(), "reservation-details", reservation)
	// We can use these models from any point of application
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	// Read flags from CLI (replace hardcoded values)
	//First flag for "production - true or false"
	InProduction := flag.Bool("production", true, "Application is in Production")
	useCache := flag.Bool("cache", true, "Use cache for templates")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPass := flag.String("dbpass", "", "Database password")
	dbPort := flag.Int("dbport", 5432, "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database SSL (disable, prefer, require")

	flag.Parse()

	//Check if all mandatory parameters are provided
	if *dbName == "" || *dbUser == "" || *dbPass == "" {
		fmt.Println("Missing mandatory flags")
		os.Exit(1)

	}

	//Create new channel for my mail chaneel
	mailChan := make(chan models.MailData)

	//Make it avaialble for other parts of package
	app.MailChan = mailChan

	//Change these to "true" when in Production
	app.InProduction = *InProduction
	// Don't use template cache (for example during dev process)
	// false for Dev, true for Prod
	app.UseCache = *useCache

	//Define new INFO and ERROR logger and make it avaialble for whole application (vial app.Infolog)
	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	//----Session managment-----------------------
	session = scs.New()
	session.Lifetime = 24 * time.Hour              //make session valid for 24 hours
	session.Cookie.Persist = true                  //Pesist session data in the cookies
	session.Cookie.SameSite = http.SameSiteLaxMode //WTF?
	session.Cookie.Secure = app.InProduction       //This is haandling https. Set true for Prod

	// Now asign whatever you have for session in main to app.Config variable
	// This will make it accessable from all part of application
	app.Session = session

	//Initialize my database connection
	log.Println("Connecting to database...")

	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPass, *dbSSL)

	db, err := driver.ConnectSQL(dsn)
	if err != nil {
		log.Fatal("Cannot connect to database. Shuttiong down...")
	}

	//----Session managment Ends------------------

	//----Tempalte cache managment-------------------
	//Call my template cache (tc) from package render
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Can't create template cache", err)
		return nil, err
	}

	// Assign my tempalte cache to configuration variable app.TemplateCache
	// This allows get cache once and do not reach it every time we browse
	app.TemplateCache = tc

	//--TEMP:Print list of all pages from tempalte cache
	fmt.Println("---This is my template cache:---")
	for pg := range tc {
		fmt.Println(pg)
	}
	fmt.Println("---End my template cache:---")

	//--TEMP ENDS:Print list of all pages from tempalte cache

	//This give render package access to our app variable
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)
	//----Tempalte cache managment Ends----------------

	// This is to create repository variable
	repo := handlers.NewRipo(&app, db)
	//Pass it back to handlers (Why?)
	handlers.NewHandlers(repo)

	return db, nil
}
