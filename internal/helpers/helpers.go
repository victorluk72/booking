package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/victorluk72/booking/internal/config"
)

//Let's get accesss to all app variables
var app *config.AppConfig

// NewHelpers sets up app config for helpers
func NewHelpers(a *config.AppConfig) {
	app = a
}

// ClientError handles the client side errors
// It take Response Writer and status code as input parameters
func ClientError(w http.ResponseWriter, status int) {

	// Write to my custom error log
	app.InfoLog.Panicln("Cient error with status of", status)

	//run standard function Error
	http.Error(w, http.StatusText(status), status)

}

// ServerError handles the server side errors
// It take Response Write and error as input parameters
// It writes the error to client browser
func ServerError(w http.ResponseWriter, err error) {

	//Get detailed information about error
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)

	//run standar function Error
	//Get internal Server error code
	ise := http.StatusInternalServerError
	http.Error(w, http.StatusText(ise), ise)

}

func IsAuthenticated(r *http.Request) bool {

	//Check if current containes key "user_id"
	exists := app.Session.Exists(r.Context(), "user_id")
	return exists
}
