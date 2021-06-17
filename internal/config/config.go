package config

import (
	"log"
	"text/template"

	"github.com/alexedwards/scs/v2"
	"github.com/victorluk72/booking/internal/models"
)

// AppConfig holds application configuration values
// This should be accessable from all ohter packages
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	InProduction  bool
	Session       *scs.SessionManager
	MailChan      chan models.MailData // CChannel for sending email
}
