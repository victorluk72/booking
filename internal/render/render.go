package render

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/justinas/nosurf"
	"github.com/victorluk72/booking/internal/config"
	"github.com/victorluk72/booking/internal/models"
)

// Define var "functions". We will use it to allow our custom functions in templates
// These will be custom functions for tempaltes (in future)
// You can define your functions in this module and pass it to templates
var functions = template.FuncMap{
	"humanDate":  HumaneDate,
	"formatDate": FormatDate,
	"iterate":    Iterate,
	"addInt":     AddInt,
}

// This variable is a pointer to my site-wide config package
var app *config.AppConfig

// This is variable to store patch to templates
// This is sfor Windows
//var pathToTemplates = "../../templates"

//This si for Linux
var pathToTemplates = "./templates"

// NewRenderer sets the config for tempalte package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// HumaneDate formats date to readable format and pass to template as a string
// Make this function available to template by putting it to var functions
func HumaneDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatDate formats any date to specified format (also passed as function argument)
// Make this function available to template by putting it to var functions
func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

// Iterate will make possible to use "for" loops in tempalte
// It takes "count" as parametes and returm slice of ints, startign from 1 to "count"
// Make this function available to template by putting it to var functions
func Iterate(count int) []int {

	var i int
	var items []int

	for i = 0; i < count; i++ {
		items = append(items, i)
	}
	return items
}

// Make this function available to template by putting it to var functions
func AddInt(a, b int) int {

	return a + b
}

// AddDefaultData is to provide default tempalte data to tempaltes
// it takes TemplateData struct as input argument and return the same struct (but with data)
// This is for data that requred in every page, so we don't need to build it on each handler
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {

	//Add CSRF token to all tempalte data
	//Token takes an HTTP request and returns
	//the CSRF token for that request or an empty string if not exist.
	td.CSRFToken = nosurf.Token(r)

	// Add flash/warning/error messages (pass through session)
	// They wil be autopopulated every time when I'm rendering pages
	// PopString put something in the session until next time page displied
	td.Flash = app.Session.PopString(r.Context(), "flash-msg")
	td.Error = app.Session.PopString(r.Context(), "error-msg")
	td.Warning = app.Session.PopString(r.Context(), "warning-msg")

	//This logic for authenticating user (we usee data from session)
	//default value for the bool type in the Go programming language is "false"
	//For authenticated user we will make it true
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuth = true
	}

	return td
}

// RenderTemplate renders the template and pass it to http.Response writer
// it accepts 4 arguments: http.ResponseWriter, http Request, name of tempalte (templ string) and td (data for template)
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {

	//If you in production mode don't use template cache, rebuild it with every request
	var tc map[string]*template.Template

	if app.UseCache == true {
		// get the tempalte cache from app.Config ( this is not reading tempalte from disc
		// but from in memory cache)
		// This is for Production mode
		tc = app.TemplateCache

	} else {
		// This will ignore tempalte cache and rebuld tempalte cache from disc every time
		// This is for development mode
		tc, _ = CreateTemplateCache()

	}

	//Pull individual template from my cache of tempaltes
	//if exist run it, otherwise error
	t, ok := tc[tmpl]
	fmt.Println("THIS IS MY TEMPLATE:", t)
	if !ok {
		log.Fatal("Couldn't get template from tempalte cache")
		return errors.New("Can't get template from cache")
	}

	//Create a bytes buffer to hold the template
	buf := new(bytes.Buffer)

	//This is to add defauld set of data to tempalte data
	td = AddDefaultData(td, r)

	//Execute template from buffer, pass td (tempalte data)
	_ = t.Execute(buf, td)

	//Write from buffer to responce writer(this basically show content)
	_, err := buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writing template to browser")
		return err
	}

	return nil

}

// CreateTemplateCache is to define map that would hold all template files
// from "template" directory index would be a file name and value is pointer to rendered tempalte
// INPORTANT: Using this map with cache prevent from reading templates from disk ever time when page loaded
// instead we read if from "in memory" cache - increase speed dramatically!
func CreateTemplateCache() (map[string]*template.Template, error) {

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
