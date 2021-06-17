package render

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/victorluk72/booking/internal/config"
	"github.com/victorluk72/booking/internal/models"
)

//variable to have ability to test render template logic
var session *scs.SessionManager
var testApp config.AppConfig

func TestMain(m *testing.M) {
	gob.Register(models.Reservation{})

	//Change these to "true" when in Production
	testApp.InProduction = false

	//Define new INFO and ERROR logger and make it avaialble for whole application (vial app.Infolog)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	testApp.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	testApp.ErrorLog = errorLog

	//----Session managment-----------------------
	session = scs.New()
	session.Lifetime = 24 * time.Hour              //make session valid for 24 hours
	session.Cookie.Persist = true                  //Pesist session data in the cookies
	session.Cookie.SameSite = http.SameSiteLaxMode //WTF?
	session.Cookie.Secure = false                  //This is haandling https. Set true for Prod

	// Now asign whatever you have for session in main to app.Config variable
	testApp.Session = session

	app = &testApp

	os.Exit(m.Run())
}

//Set up interface for http.ResponseWriter

type myWriter struct{}

//Create methods for this type that satisfy the original http.ResponseWriter
func (tw *myWriter) Header() http.Header {
	var h http.Header
	return h

}

func (tw *myWriter) WriteHeader(i int) {}

func (tw *myWriter) Write(b []byte) (int, error) {
	length := len(b)
	return length, nil
}
