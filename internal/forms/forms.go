package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Form defines form data
type Form struct {
	url.Values
	Errors errors
}

// New creates a new Form with the given data and an empty error map.
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Required checks if the specified fields are present in the form data and not empty.
// If any field is empty, an error message is added to the form indicating that the field
// cannot be blank.
func (f *Form) Required(fields ...string) {
	// For each field in the fields slice
	for _, field := range fields {

		// Get the field value
		value := f.Get(field)

		// If, after removing spaces on both sides, the field value is empty,
		// add an error message to the form. This required field was not filled.
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// Has returns true if the specified field is present in the form data,
// false otherwise.
func (f *Form) Has(field string) bool {
	// Get the form value
	x := f.Get(field)

	// If there is no form value, return false
	return x != ""
}

// Valid returns true if there are no errors in the form, false otherwise.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// MinLength checks if the specified field is present in the form data and if its length is greater
// or equal to the given minimum length. If the field value is empty or its length is less than the
// minimum length, an error message is added to the form.
func (f *Form) MinLength(field string, length int) bool {
	// Get the form value
	x := f.Get(field)

	// If the length of the form value is less than the minimum length,
	// add an error message to the form
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("Must be at least %d characters long", length))
		return false
	}

	return true
}

// IsEmail checks if the specified field is present in the form data and if it is a valid email address.
// If the field value is not a valid email address, an error message is added to the form.
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
	}
}
