package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/victorluk72/booking/internal/config"
	"github.com/victorluk72/booking/internal/models"
	"github.com/victorluk72/booking/internal/render"
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
// With the receiver for functiom (m *Repository) all my handlers has
// access to all variable from Repository
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "home.page.html", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	//Perform some busness logic and pass data to template
	stringMap := make(map[string]string)
	stringMap["testKey"] = "Sent from handler"

	render.RenderTemplate(w, r, "about.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}

// Reservation renders the reservsation page and display form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "make-reservation.page.html", &models.TemplateData{})
}

// Avaialbility renders the search-avaialibility page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "search-availability.page.html", &models.TemplateData{})
}

// PostAvaialbility handles the posted data from search-avaialibility page
// It get the data and trore them in variables
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {

	//These are data send from form
	start_date := r.Form.Get(("start_date"))
	end_date := r.Form.Get(("end_date"))

	//This is to manage data form form
	w.Write([]byte(fmt.Sprintf("Start date is: %s and end date is %s", start_date, end_date)))
}

//This is struct for our JSON data to use for AJAX to check room avaialability
type jsonResponce struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// AvailabilityJSON handles request for availability and sends JSON responce (via AJAX)
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {

	//set default JSON responce
	resp := jsonResponce{
		OK:      false,
		Message: "Available for you!",
	}

	//Marshal my struct to JSON
	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		log.Println(err)
	}

	//Create a Header for responce with JSON
	w.Header().Set("Content-Type", "application/json")

	//Write marshalled JSON to writer (display on page)
	w.Write(out)
}

// Generals renders the generals page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "generals.page.html", &models.TemplateData{})
}

//Majors renders the major page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "majors.page.html", &models.TemplateData{})
}

//Contact renders the major page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.html", &models.TemplateData{})
}
