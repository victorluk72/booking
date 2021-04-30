package handlers

import (
	"net/http"

	"github.com/victorluk72/booking/pkg/config"
	"github.com/victorluk72/booking/pkg/models"
	"github.com/victorluk72/booking/pkg/render"
)

//----This is section about Repository--------------
//variable of type Repository(we using pointer since it is struct)
var Ripo *Repository

//This is struct for repository
type Repository struct {
	App *config.AppConfig
}

// NewRipo creates  new repository
func NewRipo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandlers sets repository to the handlers
func NewHandlers(r *Repository) {
	Ripo = r

}

//----End of the section about Repository--------------

// Home is the handler for the home page
// With the receiver for functiom (m *Repository) all my hadlers has
// access to all variable from Repository
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "home.page.html", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	//Perform some busness logic and pass data to template
	stringMap := make(map[string]string)
	stringMap["testKey"] = "Sent from handler"

	render.RenderTemplate(w, "about.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}
