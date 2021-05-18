package forms

//declare package level variable to keep errors as a map structure
type errors map[string][]string

// Add is adding new errors to my "errors" map
// It accepts two parameters -fields from form and error message we want to store
func (e errors) Add(field, message string) {

	e[field] = append(e[field], message)
}

// Get return the first error message from our map
// if dosn exist return emplty string
func (e errors) Get(field string) string {

	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}
