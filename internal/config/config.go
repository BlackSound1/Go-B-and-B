package config

import (
	"html/template"
	"log"

	"github.com/BlackSound1/Go-B-and-B/internal/models"
	"github.com/alexedwards/scs/v2"
)

// AppConfig holds the application config
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	InProduction  bool
	Session       *scs.SessionManager
	MailChan      chan models.MailData
	EnvVars       map[string]any
}
