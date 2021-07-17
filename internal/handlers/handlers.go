package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

//---------------HANDLERS FOR FRONT END---------------------------------

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

// Login the handler for the login page
func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {

	//Perform some busness logic and pass data to template
	stringMap := make(map[string]string)
	stringMap["testKey"] = "Sent from handler"

	render.Template(w, r, "login.page.html", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostLogin handles login user
func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {

	//Prevent session fixation attack - renews the tocken
	_ = m.App.Session.RenewToken(r.Context())

	//Parse form
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	//Get data from form to variable
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	//Check if form is valid
	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		//If from is not valid redirect back to initial form
		render.Template(w, r, "login.page.html", &models.TemplateData{
			Form: form,
		})
		return

	}

	//If form is valid try to authenticate the user
	//Call out custm build function Authenticate that returns three parameters
	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println("Can't login", err)

		//Put error to the session, then redirect user back to page
		m.App.Session.Put(r.Context(), "error-msg", "Invalid loging credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return

	}

	//For succesfuly authenticated user store their ID in the session
	//Add it to the current session, then redirect to home page
	//See helper function IsAuthenticated()
	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash-msg", "Logged in succesfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

// Logout handles the logout logic
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {

	//Simple way to log out is to destroy the session and redirect to login page
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)

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

//---------------HANDLERS FOR ADMIN-------------------------------------
// AdminDashboard handles admin dashboard page
func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {

	//get all reservations from DB
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return

	}

	//Put my reservation to variable Data and pass to template
	data := make(map[string]interface{})

	//Asign key to my reservations data
	data["reservations"] = reservations

	render.Template(w, r, "reservations-all.page.html", &models.TemplateData{
		Data: data,
	})

}

// AdminNewReservations handles admin area - new reservation list
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {

	//get all reservations from DB
	reservations, err := m.DB.NewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return

	}

	//Put my reservation to variable Data and pass to template
	data := make(map[string]interface{})

	//Asign key to my reservations data
	data["reservations"] = reservations

	render.Template(w, r, "reservations-all.page.html", &models.TemplateData{
		Data: data,
	})
}

// AdminAllReservations handles admin area - all reservations list
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {

	//get all reservations from DB
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return

	}

	//Put my reservation to variable Data and pass to template
	data := make(map[string]interface{})

	//Asign key to my reservations data
	data["reservations"] = reservations

	render.Template(w, r, "reservations-all.page.html", &models.TemplateData{
		Data: data,
	})

}

// AdminShowReservation shows single reservation details
func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {

	//get room id and source page from URL (r.RequestURI)
	//my URLL wil look like "/admin/reservations/all/6", so I want to split it
	//You can split by position - in our case id is in position 4
	explodedURL := strings.Split(r.RequestURI, "/")

	//Get the ID from URL
	id, err := strconv.Atoi(explodedURL[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//Get the source from URL
	src := explodedURL[3]

	//Build the string map to hold source value
	stringMap := make(map[string]string)
	stringMap["src"] = src

	//get reservation from database
	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//Build the the map to hold model "reservation"
	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "admin-reservation.page.html", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})

}

// AdminCalendar displays reservation calendar
func (m *Repository) AdminCalendar(w http.ResponseWriter, r *http.Request) {

	//Assume that there is no year or month specify in URL (show current year and current month)
	now := time.Now()

	//Check if there is parameters in URL for dates ( e.g. ?y=2021&m=6)
	//if parameters exists make the year and month from url param string
	if r.URL.Query().Get("y") != "" {

		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))

		//reset now as per parameters from URL
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	}

	//Prepare data for current date
	data := make(map[string]interface{})
	data["now"] = now

	//Create the "next month" and "previouse month" button
	next := now.AddDate(0, 1, 0)
	prev := now.AddDate(0, -1, 0)

	//Format month and year
	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	prevMonth := prev.Format("01")
	prevMonthYear := prev.Format("2006")

	//Prepare data for passing to template
	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_year"] = nextMonthYear

	stringMap["prev_month"] = prevMonth
	stringMap["prev_year"] = prevMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_year"] = now.Format("2006")

	//We need to calculate number of days in any given month
	// Get the first nad last day of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	//This is how identify first day of the month
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	//This is how to identify last day of any month
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	//Create map to store number of days of each month (as integer)
	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	//Get all rooms "(pass as models)
	rooms, err := m.DB.GetAllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return

	}

	//Add rooms to the data object
	data["rooms"] = rooms

	//Build the datya structure (maps) to store information about
	//reservation and blocked rooms

	for _, x := range rooms {

		//Create out data structure (maps)
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		//Loop through each nonth (from first to last day)
		//This is how you loop through dates
		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {

			//initialize each day and make it = 0
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0

		}

		//now get all restictions of each room (mark the rooms that has a reservations)
		restrictions, err := m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		//Now loop through reservarion
		for _, y := range restrictions {

			if y.ReservationID > 0 {
				//you have reservation agains this room
				for d := y.StartDate; d.After(y.EndDate) == false; d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = y.ReservationID
				}

			} else {

				//you have owner block agains this room
				blockMap[y.StartDate.Format("2006-01-2")] = y.ID
			}
		}

		//Create data structure for hte reservations and block
		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		//Store above stucture in Session
		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)

	}

	render.Template(w, r, "admin-calendar.page.html", &models.TemplateData{
		StringMap: stringMap,
		IntMap:    intMap,
		Data:      data,
	})

}

// AdminShowReservation shows single reservation details
func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {

	//get room id and source page from URL (r.RequestURI)
	//my URLL wil look like "/admin/reservations/all/6", so I want to split it
	//You can split by position - in our case id is in position 4
	explodedURL := strings.Split(r.RequestURI, "/")

	//Get the ID from URL
	id, err := strconv.Atoi(explodedURL[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//Get the source (all or new) from URL
	src := explodedURL[3]

	//Build the string map to hold source value
	stringMap := make(map[string]string)
	stringMap["src"] = src

	//Save changes from form
	m.saveReservationDetails(id, w, r)

	//redirect to the original page (either "all" or "new")
	m.App.Session.Put(r.Context(), "flash", "Reservation updated")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)

}

// AdminProcessReservation marks reservation as processed (change status "processed = 1")
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {

	//Get the ID from URL
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//Get the source from URL
	src := chi.URLParam(r, "src")

	//m.saveReservationDetails(id, w, r)
	//This doesn't work becase can't see the form

	//Now call DB function UpdateProcessedForReservation()
	err = m.DB.UpdateProcessedForReservation(id, 1)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//Inform customer and redirect to all reservation (based on src)
	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

// AdminDeleteReservation deletess reservation
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {

	//Get the ID from URL
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//Get the source from URL
	src := chi.URLParam(r, "src")

	//Now call DB function UpdateProcessedForReservation()
	err = m.DB.DeleteReservation(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//Inform customer and redirect to all reservation (based on src)
	m.App.Session.Put(r.Context(), "flash", "Reservation deleted")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

// saveReservationDetails updates reservation from form (use in two places above)
func (m *Repository) saveReservationDetails(id int, w http.ResponseWriter, r *http.Request) {

	//Parse form
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
	}

	//Get the reservatiom we want to update by ID (from URL)
	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//Now update my modle from what is in the form
	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	log.Println("HEEREEE:", r.Form.Get("first_name"))

	//Now update table in database
	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}
