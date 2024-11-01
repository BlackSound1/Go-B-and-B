package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/BlackSound1/Go-B-and-B/internal/config"
	"github.com/BlackSound1/Go-B-and-B/internal/models"
	"github.com/justinas/nosurf"
)

var app *config.AppConfig
var functions = template.FuncMap{
	"humanDate":  HumanDate,
	"formatDate": FormatDate,
	"iterate":    Iterate,
	"add":        Add,
}
var pathToTemplates = "./templates"

// NewRenderer sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// Add returns the sum of a and b.
func Add(a, b int) int {
	return a + b
}

// Iterate returns a slice of integers. The size of the slice
// is determined by the count parameter. The slice can be used in a template to loop over the
// items in the slice. E.g. {{range .Items | Iterate 5}}<li>List item #{{.}}</li>{{end}}
func Iterate(count int) []int {
	var i int
	var items []int

	for i = 0; i < count; i++ {
		items = append(items, i)
	}

	return items
}

// FormatDate takes a time.Time and a string format and returns a string
// representing the date in the given format. It exists to be used as a
// template function in the template.FuncMap.
func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

// HumanDate takes a time.Time and returns a string representing the date in the
// format YYYY-MM-DD. It exists to be used as a template function in the
// template.FuncMap, which is why it is exposed as a public function.
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// AddDefaultData adds data to the template data that is present on every page, such
// as the CSRF token and any flash messages. It also removes them from the session
// so that they are not present on the next request.
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	// PopString puts a string into the session until the next request
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")

	// Check if user is authenticated
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}

	td.CSRFToken = nosurf.Token(r)

	return td
}

// Template renders a template to the http.ResponseWriter, given the name of the
// template and data to pass in. If app.UseCache is true, it will use the
// template cache stored in the app config. Otherwise, it will create a new cache
// with the CreateTemplateCache function. If the template does not exist in the
// cache, it will return an error. It will also add default data to the template
// data before rendering it, such as the CSRF token and any flash messages.
func Template(w http.ResponseWriter, r *http.Request, tmpl string, templateData *models.TemplateData) error {
	var templateCache map[string]*template.Template

	// If we are using the cache, get the template cache from the app config.
	// Otherwise, create a new template cache
	if app.UseCache {
		templateCache = app.TemplateCache
	} else {
		templateCache, _ = CreateTemplateCache()
	}

	// Get the template from the cache
	template, ok := templateCache[tmpl]
	if !ok {
		return errors.New("can't get template from cache")
	}

	// Create a buffer of bytes
	buffer := new(bytes.Buffer)

	templateData = AddDefaultData(templateData, r)

	_ = template.Execute(buffer, templateData)

	// Render the template
	_, err := buffer.WriteTo(w)

	if err != nil {
		log.Println("error writing template to browser:", err)
		return err
	}

	return nil

}

// CreateTemplateCache creates a map of template names to template sets.
// It will get all of the files named *.page.tmpl from the pathToTemplates
// directory and create a template with the name of the file. It will then
// attempt to find a layout template and add it to the template set. The
// template set will be added to the map with the name of the page as the key.
// It returns the populated map.
func CreateTemplateCache() (map[string]*template.Template, error) {
	// myCache := make(map[string]*template.Template)
	myCache := map[string]*template.Template{} // Same as above

	// Get all of the files named *.page.tmpl from ./templates
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))

	if err != nil {
		return myCache, err
	}

	// Range through all files ending with *.page.tmpl
	for _, page := range pages {
		// Get the file name
		name := filepath.Base(page)

		// Create a template with the name of the page
		// Associate the template with the name of the page
		templateSet, err := template.New(name).Funcs(functions).ParseFiles(page)

		if err != nil {
			return myCache, err
		}

		// Get the layout
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))

		if err != nil {
			return myCache, err
		}

		// If there is a layout
		if len(matches) > 0 {
			templateSet, err = templateSet.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))

			if err != nil {
				return myCache, err
			}
		}

		// Add the templateSet to the map
		myCache[name] = templateSet
	}

	return myCache, nil
}
