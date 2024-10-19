package render

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/BlackSound1/Go-B-and-B/internal/config"
	"github.com/BlackSound1/Go-B-and-B/internal/models"
	"github.com/justinas/nosurf"
)

var app *config.AppConfig

// NewTemplates sets the config for the template package
func NewTemplates(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	// PopString puts a string into the session until the next request
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	
	td.CSRFToken = nosurf.Token(r)

	return td
}

func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, templateData *models.TemplateData) {
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
		log.Fatal("Could not get template from template cache")
	}

	// Create a buffer of bytes
	buffer := new(bytes.Buffer)

	templateData = AddDefaultData(templateData, r)

	_ = template.Execute(buffer, templateData)

	// Render the template
	_, err := buffer.WriteTo(w)

	if err != nil {
		log.Println("error writing template to browser:", err)
	}

}

func CreateTemplateCache() (map[string]*template.Template, error) {
	// myCache := make(map[string]*template.Template)
	myCache := map[string]*template.Template{} // Same as above

	// Get all of the files named *.page.tmpl from ./templates
	pages, err := filepath.Glob("./templates/*.page.tmpl")

	if err != nil {
		return myCache, err
	}

	// Range through all files ending with *.page.tmpl
	for _, page := range pages {
		// Get the file name
		name := filepath.Base(page)

		// Create a template with the name of the page
		// Associate the template with the name of the page
		templateSet, err := template.New(name).ParseFiles(page)

		if err != nil {
			return myCache, err
		}

		// Get the layout
		matches, err := filepath.Glob("./templates/*.layout.tmpl")

		if err != nil {
			return myCache, err
		}

		// If there is a layout
		if len(matches) > 0 {
			templateSet, err = templateSet.ParseGlob("./templates/*.layout.tmpl")

			if err != nil {
				return myCache, err
			}
		}

		// Add the templateSet to the map
		myCache[name] = templateSet
	}

	return myCache, nil
}
