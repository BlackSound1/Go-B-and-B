package forms

type errors map[string][]string

// Add appends an error message to the list of errors for a specific field.
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get retrieves the first error message for a given field, or an empty string if
// there are no errors associated with the field.
func (e errors) Get(field string) string {
	es := e[field] // Get the error string

	// If no errors, return empty string
	if len(es) == 0 {
		return ""
	}

	// Otherwise, return the first error
	return es[0]
}
