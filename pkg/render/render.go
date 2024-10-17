package render

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/BlackSound1/Go-B-and-B/pkg/config"
	"github.com/BlackSound1/Go-B-and-B/pkg/models"
)

var app *config.AppConfig

// NewTemplates sets the config for the template package
func NewTemplates(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData) *models.TemplateData {

	return td
}

func RenderTemplate(w http.ResponseWriter, tmpl string, templateData *models.TemplateData) {
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

	templateData = AddDefaultData(templateData)

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
