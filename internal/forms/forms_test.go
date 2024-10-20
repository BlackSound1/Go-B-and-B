package forms

import (
	"net/url"
	"testing"
)

// TestForm_Valid tests the Valid function to ensure it returns true
// when the form data is valid
func TestForm_Valid(t *testing.T) {
	postData := url.Values{}

	// Create new Form
	form := New(postData)

	// Check if form is valid
	isValid := form.Valid()
	if !isValid {
		t.Error("got invalid when should have been valid")
	}
}

// TestForm_Required tests the Required method of the Form type.
// It creates test cases to check if the method works correctly for non-existent fields,
// fields with values, and fields with empty values.
func TestForm_Required(t *testing.T) {
	postData := url.Values{}

	// Create new Form
	form := New(postData)

	// Specify required fields
	form.Required("a", "b", "c")

	// Check if form is valid
	if form.Valid() {
		t.Error("form shows valid when it has no required fields")
	}

	// Create post data
	postData = url.Values{}

	postData.Add("a", "a")
	postData.Add("b", "b")
	postData.Add("c", "c")

	// Create new Form
	form = New(postData)

	form.Required("a", "b", "c")

	// Check if form is valid
	if !form.Valid() {
		t.Error("form shows not valid when it has required fields")
	}
}

// TestForm_Has tests the Has method of the Form type.
// It creates test cases to check if the method works correctly for non-existent fields,
// and fields that are present in the form data.
func TestForm_Has(t *testing.T) {
	postData := url.Values{}

	// Create new Form
	form := New(postData)

	// Check if form has field
	if form.Has("a") {
		t.Error("form has field when it doesn't")
	}

	// Create post data
	postData = url.Values{}

	postData.Add("a", "a")

	// Create new Form
	form = New(postData)

	// Check if form has field
	if !form.Has("a") {
		t.Error("form doesn't have field when it should")
	}
}

// TestForm_MinLength tests the MinLength method of the Form type.
// It creates test cases to check if the method works correctly for non-existent fields,
// fields with lengths less than the minimum, and fields with lengths greater than or equal to the minimum.
func TestForm_MinLength(t *testing.T) {
	postData := url.Values{}

	// Create new Form
	form := New(postData)

	// Check length of non-existant field
	form.MinLength("x", 5)

	if form.Valid() {
		t.Error("form shows min length for non-existant field")
	}

	// Check for prescence of error
	isError := form.Errors.Get("x")

	if isError == "" {
		t.Error("should have an error, but did not get one")
	}

	// Create post data
	postData = url.Values{}
	postData.Add("some_field", "some_value")

	form = New(postData)

	// Check the length of field we know isn't long enough
	form.MinLength("some_field", 100)

	if form.Valid() {
		t.Error("form shows min length is met when data is short")
	}

	// Empty field
	postData = url.Values{}
	postData.Add("some_other_field", "XYX123")

	form = New(postData)

	// Check length of field we know is long enough
	form.MinLength("some_other_field", 1)

	if !form.Valid() {
		t.Error("form shows min length is not met when data is long")
	}

	// Check for prescence of error
	isError = form.Errors.Get("some_other_field")

	if isError != "" {
		t.Error("should not have an error, but did get one")
	}
}

// TestForm_IsEmail tests the IsEmail method of the Form type.
// It creates a few test cases to check if the method works correctly
// for non-existent fields, valid email addresses, and invalid email addresses.
func TestForm_IsEmail(t *testing.T) {
	postData := url.Values{}

	// Create new Form
	form := New(postData)

	// Check if email is valid but for non-existent field
	form.IsEmail("email")

	if form.Valid() {
		t.Error("form shows valid email for non-existent field")
	}

	postData = url.Values{}

	// Add valid email
	postData.Add("email", "hey@hey.hey")

	form = New(postData)

	// Check if email is valid
	form.IsEmail("email")

	if !form.Valid() {
		t.Error("form shows invalid email for valid email address")
	}

	postData = url.Values{}

	// Add invalid email
	postData.Add("email", "hey")

	form = New(postData)

	// Check if email is invalid
	form.IsEmail("email")

	if form.Valid() {
		t.Error("form shows valid email for invalid email address")
	}
}
