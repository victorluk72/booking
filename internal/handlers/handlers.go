package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/victorluk72/booking/internal/config"
	"github.com/victorluk72/booking/internal/forms"
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

	//Display content from model Reservation (empty from begining)
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	//This is to render empty form (first time click on route)
	render.RenderTemplate(w, r, "make-reservation.page.html", &models.TemplateData{

		//Include empty form (see package forms)
		Form: forms.New(nil),
		Data: data, //These are data from previouse form or blank
	})
}

// PostReservation handles the post request from Reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {

	//Try to parse form first, see if thre is any arror in general
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}

	// This variable stores all data from form in the struct
	// Use model Reservation form package models)
	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	// Make a new form and pass date from Post request
	form := forms.New(r.PostForm)

	//------This is my server site form validation rules-----
	//Does this form has values in provided fields
	form.Required("first_name", "last_name", "email")

	//Does fiels matches minimum character count?
	form.MinLength("first_name", 2, r)

	//Does email field value is in proper format?
	form.IsEmail("email")

	//------End of the server site form validation rules-----

	// Check if form is NOT valid (use function from package "forms")
	if !form.Valid() {

		//This is for not Valis form
		log.Println("Form is not valid")

		// Create set of data to pass to the form is not valid
		// This structure will be sent just to keep what user already entered
		// We will repopulate these data to reloaded form
		data := make(map[string]interface{})
		data["reservation"] = reservation

		//Now render the form and pass all stored data
		render.RenderTemplate(w, r, "make-reservation.page.html", &models.TemplateData{

			//Pass form and the data that user entered
			Form: form,
			Data: data, //these are data saved from initial user entering action
		})

		//Stop here since form is not valid
		return

	}

	//----This is the "happy path" when form is valid
	// We use session for exchangind data between two pages
	m.App.Session.Put(r.Context(), "reservation-details", reservation)

	// Redirect to page "reservation-summary" with redirect status 303
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)

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

// Majors renders the major page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "majors.page.html", &models.TemplateData{})
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.html", &models.TemplateData{})
}

// ReservationSummary renders the  reservation-summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	//Get data from form into variable (use sesssion Contaxt to get it from)
	reservation, ok := m.App.Session.Get(r.Context(), "reservation-details").(models.Reservation)

	//This is to show when page open without data from form
	//It might happend when user went to page "reservation-summary" not from page with form
	if !ok {
		log.Println("Can't get item from session")
		//Put some message to the session
		m.App.Session.Put(r.Context(), "error-msg", "Can't get reservation from session")

		//Now redirect to hope page with redirect status 307
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//Clear the session, we already grabbed data (as variable "reservation")
	//Now we want to clear these data from session
	m.App.Session.Remove(r.Context(), "reservation-details")

	//Make a map to pass data from form
	data := make(map[string]interface{})

	//Create a map with index "reservation" and value as struct Reservation
	data["i-reservation"] = reservation

	render.RenderTemplate(w, r, "reservation-summary.page.html", &models.TemplateData{
		Data: data,
	})
}
