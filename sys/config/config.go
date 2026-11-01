// Package config reads the configuration of the app with configloader: the
// defaults below, then conf.json (optional), then APP_* environment variables
// and then command line flags, each one beating the one before. If anything is
// wrong Load says all of it and main refuses to start.
package config

import (
	"errors"
	"log/slog"
	"os"

	"github.com/deltegui/configloader/v3"
)

type Configuration struct {
	Application struct {
		// Name is the name of the app: the banner, the title and the top bar.
		Name string `config:",required" usage:"name of the application"`
		// Version is set by the code, not by the configuration.
		Version string `config:"-"`
	}
	// ListenURL is APP_ADDR or -addr.
	ListenURL string `config:"addr" default:":8080" usage:"address to listen on"`
	// Secret seals cookies and CSRF tokens: 32 bytes. `make init` makes one.
	Secret []byte `config:",required,secret" usage:"32 bytes that seal cookies and CSRF tokens"`
	// CookieSecure needs https; turn it off only for development over plain http.
	CookieSecure bool `default:"true"`
	Logger       struct {
		// Level is debug, info, warn or error (anything slog.Level reads).
		Level slog.Level `default:"info"`
	}
	Database struct {
		URL string `config:",required,secret" usage:"postgres connection URL"`
	}
	// Admin makes the first admin when there are no users.
	Admin struct {
		Email    string
		Password string `config:",secret"`
	}
}

// Validate is what a single field cannot check.
func (c Configuration) Validate() error {
	var errs []error
	if len(c.Secret) != 32 {
		errs = append(errs, errors.New("Secret must be exactly 32 bytes"))
	}
	if (c.Admin.Email == "") != (c.Admin.Password == "") {
		errs = append(errs, errors.New("Admin.Email and Admin.Password go together: set both or none"))
	}
	return errors.Join(errs...)
}

// Load returns the configuration or an error that lists every problem. It does
// not exit: the program decides.
func Load() (Configuration, error) {
	var conf Configuration
	conf.Application.Version = "0.1.0"
	err := configloader.Load(&conf,
		configloader.OptionalFile("conf.json"),
		configloader.Env("APP"),
		configloader.Flags(os.Args[1:]),
	)
	return conf, err
}
