package main

import (
	"fmt"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/victorluk72/booking/internal/config"
)

func TestRoutes(t *testing.T) {

	//Let's test if routes return proper type (should be mux)
	var app config.AppConfig
	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
		fmt.Printf("My type is %T", v)
		//case http.Handler:
		//fmt.Printf("My type is %T", v)
		//do nothing you pass
	default:
		t.Error(fmt.Sprintf("type is not *chi.Mux type, but %T", v))
	}

}
