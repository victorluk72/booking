package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/victorluk72/booking/internal/config"
	"github.com/victorluk72/booking/internal/driver"
	"github.com/victorluk72/booking/internal/forms"
	"github.com/victorluk72/booking/internal/helpers"
	"github.com/victorluk72/booking/internal/models"
	"github.com/victorluk72/booking/internal/render"
	"github.com/victorluk72/booking/internal/repository"
	"github.com/victorluk72/booking/internal/repository/dbrepo"
)

//----This is section about Repository--------------
//variable of type Repository(we using pointer since it is struct)
var Ripo *Repository

//This is struct for repository. This help to have access to application level stuff
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRipo creates new repository
func NewRipo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewHandlers sets repository to the handlers
func NewHandlers(r *Repository) {
	Ripo = r

}

//----End of the section about Repository--------------

// Home is the handler for the home page
// With the receiver for functiom (m *Repository) all my handlers has
// access to all variable from Repository (app configs and DB access in particular)
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.html", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	//Perform some busness logic and pass data to template
	stringMap := make(map[string]string)
	stringMap["testKey"] = "Sent from handler"

	render.Template(w, r, "about.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}

// Loginis the handler for the login page
func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {

	//Perform some busness logic and pass data to template
	stringMap := make(map[string]string)
	stringMap["testKey"] = "Sent from handler"

	render.Template(w, r, "login.page.html", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// Reservation renders the reservsation page and display form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {

	//Get my reservation from session and put into variable, convert to string
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("cannot get reservation data from session"))
		return
	}

	//Get my room details using custom func GetRoomByID
	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		//Use our custom built ServerError helper
		helpers.ServerError(w, err)
		return
	}

	//Store room details in my res variable (which represent model Reservation)
	res.Room.RoomName = room.RoomName

	//Put my reservation into session
	m.App.Session.Put(r.Context(), "reservation", res)

	//Cast dates back to string for start and end dates
	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	//Pass as a string via model TemplateData.StringMap
	stringMapDates := make(map[string]string)
	stringMapDates["start_date"] = sd
	stringMapDates["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	//This is to render empty form (first time click on route)
	render.Template(w, r, "make-reservation.page.html", &models.TemplateData{

		//Include empty form (see package forms)
		Form:      forms.New(nil),
		Data:      data,           //These are data from previouse form or blank
		StringMap: stringMapDates, //This to pass dates as string
	})
}

// PostReservation handles the post request from Reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {

	//Pull my reservation model from session
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("Cannot get reservation model from session"))
		return
	}

	//Try to parse form first, see if thre is any arror in general
	err := r.ParseForm()

	if err != nil {
		//Use our custom built ServerError helper
		helpers.ServerError(w, err)
		return
	}

	//--WORK WITH DB-------------------------------------

	// This variable stores all data from form in the struct
	// Use model Reservation form package models)
	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

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
		render.Template(w, r, "make-reservation.page.html", &models.TemplateData{

			//Pass form and the data that user entered
			Form: form,
			Data: data, //these are data saved from initial user entering action
		})

		//Stop here since form is not valid
		return

	}

	//Add the info from form to database
	newReservationID, err := m.DB.InsertReservstion(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//Send updated reservation model to the session
	m.App.Session.Put(r.Context(), "reservation", reservation)

	//You also nee to add new reservation to RoomRestrition table
	//Build model for RoomRestriction table
	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		ReservationID: newReservationID,
		RestrictionID: 1,
		RoomID:        reservation.RoomID,
	}

	//Add the info  to database
	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//--WORK WITH DB ENDS HERE-------------------------------------

	//--SENDING EMAIL NOTIFICATIONS-------------------------------------

	// 1) Send email to guest first

	//Build the content here as HTML string
	htmlMessage := fmt.Sprintf(`<strong>Your reservation has been completed</strong><br>
	               Dear %s,<br>
				   This is to confirm your reservation from %s ti %s.
	             `, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	//Build the message
	msg := models.MailData{
		To:      reservation.Email,
		From:    "noreply@server.com",
		Subject: "Your reservation is received",
		Content: htmlMessage,
	}

	//Pass message to channel. This will send email in background (asyncronically)
	m.App.MailChan <- msg

	//--SENDING EMAIL NOTIFICATIONS ENDS HERE---------------------------

	//----This is the "happy path" when form is valid
	// We use session for exchangind data between two pages
	m.App.Session.Put(r.Context(), "reservation-details", reservation)

	// Redirect to page "reservation-summary" with redirect status 303
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)

}

// Avaialbility renders the search-avaialibility page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.html", &models.TemplateData{})
}

// PostAvaialbility handles the posted data from search-avaialibility page
// It get the data and trore them in variables
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {

	//These are data sent from form. They come in string type, need to convert to time.Time
	start_date := r.Form.Get(("start_date"))
	end_date := r.Form.Get(("end_date"))

	//format I'm getting my dates in from the form
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start_date)
	if err != nil {
		//show error to browser
		helpers.ServerError(w, err)
	}
	endDate, err := time.Parse(layout, end_date)
	if err != nil {
		//show error to browser
		helpers.ServerError(w, err)
	}

	//Call my database function
	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		//show error to browser
		helpers.ServerError(w, err)
		return
	}

	for _, i := range rooms {
		m.App.InfoLog.Println("ROOM:", i.ID, i.RoomName)

	}

	//check if any room is avaialble
	if len(rooms) == 0 {
		//this is logic for no rooms avaialble
		//Generate error message when no rooms available
		m.App.Session.Put(r.Context(), "error-msg", "No rooms avalable for these dates")

		//redirect to the same page
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return

	}

	//Prepare date to pass available rooms ot tempalte
	data := make(map[string]interface{})
	data["rooms"] = rooms

	// We want to store data about startData, endDate and pass it in th session
	// We will use it for Make Reservation page as default data
	// To do so we would create empty Reservstion model, populate start, end Date
	// and pass to session
	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	//Put my reservation to the session under name "reservation"
	// It wil be available when we get to "Make reservation" page
	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "rooms.page.html", &models.TemplateData{
		Data: data,
	})

}

//This is struct for our JSON data to use for AJAX to check room avaialability
type jsonResponce struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles request for availability and sends JSON responce (via AJAX)
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {

	//Get start and end dates from the form and convert to time.Time format
	sd := r.Form.Get("start")
	ed := r.Form.Get("end")
	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	//format I'm getting my dates in from the form
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		//show error to browser
		helpers.ServerError(w, err)
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		//show error to browser
		helpers.ServerError(w, err)
	}

	//Check for availability by room id (use custom function SearchAvailabilityByDatesByRoomID)
	//It returns boolean value and error
	avaialable, _ := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)

	//set default JSON responce
	resp := jsonResponce{
		OK:        avaialable,
		Message:   "",
		RoomID:    strconv.Itoa(roomID),
		StartDate: sd,
		EndDate:   ed,
	}

	//Marshal my struct to JSON
	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		//Use our custom built ServerError helper
		helpers.ServerError(w, err)
		return
	}

	//Create a Header for responce with JSON
	w.Header().Set("Content-Type", "application/json")

	//Write marshalled JSON to writer (display on page)
	w.Write(out)
}

// Generals renders the generals page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.html", &models.TemplateData{})
}

// Majors renders the major page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.html", &models.TemplateData{})
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.html", &models.TemplateData{})
}

// ReservationSummary renders the reservation-summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	//Get data from form into variable (use sesssion Contaxt to get it from)
	reservation, ok := m.App.Session.Get(r.Context(), "reservation-details").(models.Reservation)

	//This is to show when page open without data from form
	//It might happend when user went to page "reservation-summary" not from page with form
	if !ok {
		m.App.ErrorLog.Println("Can't get error from session")
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

	//Format our start and end dates to string format
	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	//Put the start and end dates to stringMap
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w, r, "reservation-summary.page.html", &models.TemplateData{
		Data:      data,      //this is to pass reservation model
		StringMap: stringMap, //this is to pass my start nad end dates
	})
}

// ChooseRoom displays list of avaialble rooms
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {

	//Read room id from the URL and store as variable
	//We use build-in chi method URLParam
	RoomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		//show error to browser
		helpers.ServerError(w, err)
		return
	}

	//Get my reservation from session and put into variable, convert to string
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, err)
		return
	}

	res.RoomID = RoomID

	//Now my variable res has three values, start and End date and room it
	//I'm adding it back to session
	m.App.Session.Put(r.Context(), "reservation", res)

	//Redirect to reservation page with redirect status 303
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}

// BookRoom parse the URL (get "id, sd and ed", build session variable
// and redirect to reservation page)
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {

	//Grab the parameters from URL (id, "s" for start date, "e" for end date)
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("sd")
	ed := r.URL.Query().Get("ed")

	//Prepare the variable to hold reservation data

	var res models.Reservation

	//format my dates from URL paramaters
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		//show error to browser
		helpers.ServerError(w, err)
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		//show error to browser
		helpers.ServerError(w, err)
	}

	//Get my room details using custom func GetRoomByID
	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		//Use our custom built ServerError helper
		helpers.ServerError(w, err)
		return
	}

	//Store room details in my res variable (which represent model Reservation)
	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate
	res.Room.RoomName = room.RoomName

	//Put res (model) to session to pass to next page
	m.App.Session.Put(r.Context(), "reservation", res)

	//Redirect to Reservation page a
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

	log.Println(roomID, startDate, endDate)

}
