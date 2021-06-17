package main

import (
	"net/http"
	"os"
	"testing"
)

//This function runs before our test runs
// 1) Do something
// 2) Run the test
// 3) Exit
func TestMain(m *testing.M) {

	//1) Do somethign

	//Run the test and exit
	os.Exit(m.Run())

}

// Custom type for handler.
// need to satisfy the function ServeHTTP from http.Handler
type myHandler struct{}

func (mh *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
