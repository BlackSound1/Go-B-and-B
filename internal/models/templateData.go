package models

import "github.com/BlackSound1/Go-B-and-B/internal/forms"

// TemplateData holds data sent from handlers to templates
type TemplateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float32
	Data            map[string]interface{}
	CSRFToken       string // For CSRF protection
	Flash           string // For flash messages
	Warning         string
	Error           string
	Form            *forms.Form
	IsAuthenticated int
}
