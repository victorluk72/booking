package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// Create data structure to hold data that we are posting from form
type postData struct {
	key   string
	value string
}

//This is slice of structs for different tests
//We define struct and asign the values
var theTests = []struct {
	name               string     //name of individual test
	url                string     //path for our routes
	method             string     //GET or POST
	params             []postData //The data that being posted
	expectedStatusCode int        //status from server (e.g. 200, 401, etc)

}{
	//These are settings for GET URLs
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	{"gn", "/generals", "GET", []postData{}, http.StatusOK},
	{"mg", "/majors", "GET", []postData{}, http.StatusOK},
	{"sa", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"mr", "/make-reservation", "GET", []postData{}, http.StatusOK},
	{"rs", "/reservation-summary", "GET", []postData{}, http.StatusOK},

	//These are settings for POST URLs
	{"post-search-avail", "/search-availability", "POST", []postData{
		{key: "start", value: "2020-01-01"},
		{key: "end", value: "2020-01-06"},
	}, http.StatusOK},

	{"post-search-avail-json", "/search-availability-json", "POST", []postData{
		{key: "start", value: "2020-01-01"},
		{key: "end", value: "2020-01-06"},
	}, http.StatusOK},

	{"post-make-res", "/make-reservation", "POST", []postData{
		{key: "first_name", value: "Tom"},
		{key: "last_name", value: "Hanks"},
		{key: "email", value: "tom@hanks.com"},
		{key: "phone", value: "455555555"},
	}, http.StatusOK},
}

// The function for test itself
func TestHandlers(t *testing.T) {

	//Get my routes from setup_test.go. It return mux
	routes := getRoutes()

	//We need to create testing web server that would return us status code
	//This server will run only during the test, listen for request and return status code
	//You need to close it after done (use defer)
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	//Now run thre the range of my slice of structs theTests
	for _, e := range theTests {

		//make different tests for GET and POST methods
		if e.method == "GET" {

			//Do stuff for GET method
			//We need to act as a client (browser) and send request to server
			//Return response and error
			resp, err := ts.Client().Get(ts.URL + e.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			//Now check is status code from response matches what we expected
			//If it don't match test fails
			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}

		} else {

			//Do stuff for POST method
			//Variable to hold data from form
			values := url.Values{}

			//Now populate the valus looping through params
			for _, x := range e.params {
				values.Add(x.key, x.value)
			}

			//Run the client and get a response from Post form
			resp, err := ts.Client().PostForm(ts.URL+e.url, values)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			//Now check is status code from response matches what we expected
			//If it don't match test fails
			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}

		}

	}

}
