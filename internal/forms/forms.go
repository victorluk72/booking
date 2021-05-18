package forms

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// This is a struct to store form values and errors
// It embeds url.Value object from standard net/url
type Form struct {
	url.Values
	Errors errors
}

// Valid is to validate if from is valid
// Return  bool (tru fo rvalid form)
func (f *Form) Valid() bool {

	//If we have error in the form, the len of error != 0, so we retrun false
	//if no errors, then length == 0, so we return true
	return len(f.Errors) == 0

}

// New initializes a Form struct
func New(data url.Values) *Form {

	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Required is to validate all mandatry fields from form
// if no value add to the map of errors!
// takes fields as variadic parameter (as many fields as you need)
func (f *Form) Required(fields ...string) {

	for _, field := range fields {
		value := f.Get(field)

		//Check if thre is any value in the field
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field can't be blank")
		}
	}

}

// Has checks if field is in Post request and not empty
// input parameters are teh filed name and request itself
// return "true" is field exist and "false" if missing
func (f *Form) Has(field string, r *http.Request) bool {
	//varible to store value from form field
	x := r.Form.Get(field)

	if x == "" {
		return false
	}
	return true
}

// MinLength checks for minimum character count for value from form field and returns bool
func (f *Form) MinLength(field string, length int, r *http.Request) bool {

	//varible to store value from form field
	x := r.Form.Get(field)

	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This filed must be at least %d characters long", length))
		return false
	}
	return true
}

// IsEmail validates if files is in proper email format
// It uses third party package Govalidator
func (f *Form) IsEmail(field string) bool {

	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
		return false
	}

	//Happy path
	return true
}
