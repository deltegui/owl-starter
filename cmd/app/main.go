// Command app is the entry point: it reads the configuration, connects to
// Postgres, applies the migrations, wires OWL and starts the server. Each
// folder of feature/ is a slice with its own bootstrap, handler and service.
package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/cipher"
	"github.com/deltegui/owl/v3/core"
	"github.com/deltegui/owl/v3/csrf"
	"github.com/deltegui/owl/v3/hash"
	"github.com/deltegui/owl/v3/logx"
	"github.com/deltegui/owl/v3/middleware"
	"github.com/deltegui/owl/v3/session"

	"github.com/deltegui/owl-starter/domain/user"
	"github.com/deltegui/owl-starter/feature/account"
	"github.com/deltegui/owl-starter/feature/home"
	"github.com/deltegui/owl-starter/feature/language"
	"github.com/deltegui/owl-starter/sys/bootstrap"
	"github.com/deltegui/owl-starter/sys/config"
	"github.com/deltegui/owl-starter/sys/persistence"
	"github.com/deltegui/owl-starter/web/layout"
	"github.com/deltegui/owl-starter/web/locale"
	"github.com/deltegui/owl-starter/web/static"
)

// banner.txt is the ASCII art printed at startup; replace it with your own.
//
//go:embed banner.txt
var banner embed.FS

var modules = []bootstrap.BootstrapFunc{
	language.Bootstrap,
	account.Bootstrap,
	home.Bootstrap,
}

func main() {
	conf, err := config.Load()
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		// Better not to start at all than to run with a configuration we do not understand.
		fmt.Fprintf(os.Stderr, "app: invalid configuration:\n%v\n", err)
		os.Exit(1)
	}
	owl.PrintLogoEmbedded(banner, "banner.txt")
	fmt.Println(conf.Application.Name, conf.Application.Version)
	layout.Name = conf.Application.Name
	// The logger needs the configuration (the level), so what fails before this
	// is written to stderr by hand.
	logger := logx.Default{Level: logx.Level(conf.Logger.Level)}
	ctx, err := wire(conf, logger)
	if err != nil {
		logger.Error("cannot start", logx.Err(err))
		os.Exit(1)
	}
	ctx.Mux.Listen(conf.ListenURL)
}

func wire(conf config.Configuration, logger logx.Logger) (bootstrap.Context, error) {
	bg := context.Background()

	db, err := persistence.Connect(conf.Database.URL)
	if err != nil {
		return bootstrap.Context{}, err
	}
	if err := persistence.Migrate(bg, db); err != nil {
		return bootstrap.Context{}, err
	}

	cy, err := cipher.NewWithKeys(conf.Secret)
	if err != nil {
		return bootstrap.Context{}, err
	}

	store := persistence.NewSessionStore(db)
	go sweepSessions(store, logger.WithModuleName("sessions"))
	sessions := session.NewManager(cy,
		session.WithStore(store),
		session.WithSecure(conf.CookieSecure))
	cs := csrf.New(cy, csrf.WithSecure(conf.CookieSecure))

	router := owl.New(cy)
	router.Logger = logger
	router.StaticEmbedded(static.Files)
	router.AddLocalization(owl.Localization{
		Shared:    locale.Shared,
		Errors:    locale.Errors,
		Languages: locale.Languages,
	})
	router.Use(middleware.Logger)
	router.Use(middleware.Csrf(cs))

	ctx := bootstrap.Context{
		Mux:           router,
		Logger:        logger,
		Configuration: conf,
		DB:            db,
		Cipher:        cy,
		Hasher:        hash.NewArgon2(),
		Sessions:      sessions,
		Csrf:          cs,
	}

	if err := createFirstAdmin(bg, ctx); err != nil {
		return bootstrap.Context{}, err
	}

	for _, bootstrapFunc := range modules {
		bootstrapFunc(ctx)
	}
	return ctx, nil
}

// createFirstAdmin makes the admin of Admin.Email and Admin.Password
// when there are no users, so a new database can be logged into.
func createFirstAdmin(ctx context.Context, btx bootstrap.Context) error {
	if btx.Configuration.Admin.Email == "" {
		return nil
	}
	btx.Logger.InfoContext(ctx, "creating admin user...", "email", btx.Configuration.Admin.Email)
	users := persistence.NewUserRepository(btx.DB)
	n, err := users.Count(ctx)
	if err != nil || n > 0 {
		return err
	}
	hashed, err := btx.Hasher.Hash(btx.Configuration.Admin.Password)
	if err != nil {
		return err
	}
	_, err = users.Create(ctx, user.User{
		Name:         "Admin",
		Email:        btx.Configuration.Admin.Email,
		PasswordHash: hashed,
		Role:         core.RoleAdmin,
	})
	if err == nil {
		btx.Logger.InfoContext(ctx, "created the first admin", "email", btx.Configuration.Admin.Email)
	}
	btx.Logger.InfoContext(ctx, "done creating admin user!")
	return err
}

func sweepSessions(store persistence.SessionStore, logger logx.Logger) {
	for range time.Tick(time.Hour) {
		if err := store.Sweep(context.Background()); err != nil {
			logger.Error("sweeping sessions", logx.Err(err))
		}
	}
}
