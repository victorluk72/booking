package main

import (
	"fmt"
	"net/http"
	"testing"
)

// test for NoSurf function
func TestNoSurf(t *testing.T) {

	//This is my custom type for handler
	var myH myHandler

	h := NoSurf(&myH)

	//If my type is htttp.Handler - test pass
	switch v := h.(type) {
	//check if my h is type of http.Handler
	case http.Handler:
		//do nothing-we pass
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, but %T", v))
	}
}

// test for SessionLoad function
func TestSessionLoad(t *testing.T) {

	//This is my custom type for handler
	var myH myHandler

	h := SessionLoad(&myH)

	//If my type is htttp.Handler - test pass
	switch v := h.(type) {
	//check if my h is type of http.Handler
	case http.Handler:
		//do nothing-we pass
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, but %T", v))
	}
}
