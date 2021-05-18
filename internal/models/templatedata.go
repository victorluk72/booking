package models

import "github.com/victorluk72/booking/internal/forms"

// Create type that would hold all data types to pass from handlers to tempalte
// Special notes:
//    Form use to pass form object from standatd library
//    Data use to pass values of the form fileds
type TemplateData struct {
	StringMap map[string]string      //this holds all possible string type data
	IntMap    map[string]int         //this holds all possible integer type data
	FloatMap  map[string]float32     //this holds all possible float type data
	Data      map[string]interface{} //this holds all possible type of data-anything
	CSRFToken string                 // string for CSRF token (for form safety)
	Flash     string                 // string for flash message
	Warning   string                 // string for warning message
	Error     string                 // string for error message
	Form      *forms.Form            // form type from standard library
}
