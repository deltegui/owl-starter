package account

import (
	"github.com/deltegui/owl-starter/sys/bootstrap"
	"github.com/deltegui/owl-starter/sys/persistence"
)

func Bootstrap(ctx bootstrap.Context) {
	service := newService(persistence.NewUserRepository(ctx.DB), ctx.Hasher)

	ctx.Mux.Get(bootstrap.LoginURL, handleShowLogin())
	ctx.Mux.Post(bootstrap.LoginURL, handleLogin(service, ctx.Sessions, ctx.Csrf))
	ctx.Mux.Post("/logout", handleLogout(ctx.Sessions))
}
