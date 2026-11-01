// Package bootstrap is what every feature gets to wire itself: the multiplexer
// and what it needs from the system. Each feature has a Bootstrap(Context) that
// registers its routes; cmd/app lists them.
package bootstrap

import (
	"database/sql"

	"github.com/deltegui/owl-starter/sys/config"

	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/core"
	"github.com/deltegui/owl/v3/csrf"
	"github.com/deltegui/owl/v3/logx"
	"github.com/deltegui/owl/v3/middleware"
	"github.com/deltegui/owl/v3/session"
)

type BootstrapFunc func(Context)

const LoginURL = "/login"

type Context struct {
	Mux           *owl.Mux
	Logger        logx.Logger
	Configuration config.Configuration
	DB            *sql.DB
	Cipher        core.Cipher
	Hasher        core.Hasher
	Sessions      *session.Manager
	Csrf          *csrf.Csrf
}

// Auth lets through only a request with a session.
func (c Context) Auth() owl.Middleware { return middleware.Authorize(c.Sessions, LoginURL) }

// Admin lets through only an admin.
func (c Context) Admin() owl.Middleware { return middleware.Admin(c.Sessions, LoginURL) }
