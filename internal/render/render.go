package render

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/justinas/nosurf"
	"github.com/victorluk72/booking/internal/config"
	"github.com/victorluk72/booking/internal/models"
)

// Define var "functions". We will use it to allow our own functions in tempalte
// These will be a custom functions for tempaltes (in future)
var functions = template.FuncMap{}

// This variable is a pointer to my site-wide config package
var app *config.AppConfig

// NewTemplates sets the config for tempalte package
func NewTemplates(a *config.AppConfig) {
	app = a
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

	return td
}

// RenderTemplate rendes the template and pass it to http.Response writes
// it accepts three arguments: http.ResponseWriter, name of tempalte (templ string) and td (data for template)
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) {

	//If you in production mode don't use template cache, rebuild it with every request
	var tc map[string]*template.Template

	if app.UseCache == true {
		// get the tempalte cache from app.Config ( this is not reading tempalte from disc
		// but from in memory cache)
		// This is for Production mode
		tc = app.TemplateCache

	} else {
		// This will ignote tempalte cache and rebuld tempalte cache from disc every time
		// This is for development mode
		tc, _ = CreateTemplateCache()

	}

	//Pull individual tempalte from my cache of tempalte
	//if exist run it, otherwise error
	t, ok := tc[tmpl]
	if !ok {
		log.Fatal("Couldn't get template from tempalte cache")
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
	}

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
	pages, err := filepath.Glob("../../templates/*.page.html")
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
		matches, err := filepath.Glob("../../templates/*.layout.html")
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("../../templates/*.layout.html")
			if err != nil {
				return myCache, err
			}
		}

		//Add template set to the myCache variable
		myCache[name] = ts
	}

	return myCache, nil
}
