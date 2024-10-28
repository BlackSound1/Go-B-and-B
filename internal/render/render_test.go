package render

import (
	"net/http"
	"testing"

	"github.com/BlackSound1/Go-B-and-B/internal/models"
)

// TestAddDefaultData tests that the flash value from the session is added to the
// TemplateData struct by the AddDefaultData function.
func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()

	if err != nil {
		t.Error(err)
	}

	// Put some flash data into the session
	session.Put(r.Context(), "flash", "123")

	result := AddDefaultData(&td, r)

	if result.Flash != "123" {
		t.Error("flash value of 123 not found in session")
	}
}

// TestRenderTemplate tests the RenderTemplate function to ensure it can render a
// template to a http.ResponseWriter and also handles a template that does not
// exist.
func TestRenderTemplate(t *testing.T) {
	// Change the path to the templates folder
	pathToTemplates = "./../../templates"

	templateCache, err := CreateTemplateCache()

	if err != nil {
		t.Error(err)
	}

	// Set the template cache in the app config
	app.TemplateCache = templateCache

	// Get the session
	r, err := getSession()

	if err != nil {
		t.Error(err)
	}

	// Create a dummy response writer
	var ww myWriter

	// Render the template
	err = Template(&ww, r, "home.page.tmpl", &models.TemplateData{})

	if err != nil {
		t.Error("error writing template to browser")
	}

	// Render a non-existant template
	err = Template(&ww, r, "xyz123.page.tmpl", &models.TemplateData{})

	if err == nil {
		t.Error("rendered template that does not exist")
	}

	// Test the branch where the UseCache is true
	app.UseCache = true

	// Render the template with UseCahce
	_ = Template(&ww, r, "home.page.tmpl", &models.TemplateData{})
}

// TestNewTempleset tests that the template cache is created when calling the NewTemplates
// function and associates it with the app config.
func TestNewTempleset(t *testing.T) {
	NewRenderer(app)
}

// TestCreateTemplateCache tests that the template cache is created when calling the
// CreateTemplateCache function.
func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "./../../templates"

	_, err := CreateTemplateCache()

	if err != nil {
		t.Error(err)
	}
}

// getSession retrieves a new HTTP request and loads the session information from the request header
func getSession() (*http.Request, error) {

	// Create a dummy request
	r, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		return nil, err
	}

	// Get the context from the request
	ctx := r.Context()

	// Load the "X-Session" header into the context
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))

	// Add the newly-populated context to the request
	r = r.WithContext(ctx)

	return r, nil
}
