package render

import (
	"net/http"
	"testing"

	"github.com/victorluk72/booking/internal/models"
)

func TestAddDefaultData(t *testing.T) {

	var td models.TemplateData

	r, err := getSession()
	if err != nil {
		t.Error("fail - AddDefaultData")
	}

	//Add something to the session and see if it pass
	session.Put(r.Context(), "flash-msg", "123")

	result := AddDefaultData(&td, r)
	if result.Flash != "123" {
		t.Error("Flash value 123 not found in session")
	}

}

func TestRenderTemplate(t *testing.T) {

	pathToTemplates = "../../templates"

	//Create template cache
	tc, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

	//Pass template cache (tc) to app level varaiable
	app.TemplateCache = tc

	// Get the response from getSession function (see below)
	r, err := getSession()
	if err != nil {
		t.Error(err)

	}

	//This is artificialy created type that satisfy inteface of http.ResponseWriter
	var rw myWriter

	err = RenderTemplate(&rw, r, "home.page.html", &models.TemplateData{})
	if err != nil {
		t.Error("Error writing template to browser")
	}

	// err = RenderTemplate(&rw, r, "wrong.page.html", &models.TemplateData{})
	// if err == nil {
	// 	t.Error("Got template that doesn't exist")
	// }

}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "../../templates"

	_, err := CreateTemplateCache()
	if err != nil {
		t.Error("Error creating template cache")
	}

}

func getSession() (*http.Request, error) {

	//Create the request
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}

	//variable for context
	ctx := r.Context()

	//Put session data into context
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))

	//add it back to request
	r = r.WithContext(ctx)
	return r, nil

}
