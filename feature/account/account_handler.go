package account

import (
	"errors"
	"net/http"

	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/core"
	"github.com/deltegui/owl/v3/csrf"
	"github.com/deltegui/owl/v3/session"

	"github.com/deltegui/owl-starter/domain/user"
	"github.com/deltegui/owl-starter/sys/bootstrap"
)

func handleShowLogin() owl.Handler {
	return func(ctx owl.Ctx) error {
		if ctx.HaveSession() {
			return ctx.RedirectCode("/", http.StatusSeeOther)
		}
		return renderLogin(ctx, loginViewModel{isInvalid: false})
	}
}

func handleLogin(svc service, sessions *session.Manager, cs *csrf.Csrf) owl.Handler {
	return func(ctx owl.Ctx) error {
		var form loginForm
		ctx.ParseForm(&form)
		ctx.Validate(form)
		if !ctx.ModelState.Valid {
			ctx.Status(http.StatusBadRequest)
			return renderLogin(ctx, loginViewModel{
				form:      form,
				isInvalid: false,
			})
		}

		u, err := svc.authenticate(ctx.Context(), form.Email, form.Password)
		if errors.Is(err, errInvalidCredentials) {
			ctx.Status(http.StatusUnauthorized)
			return renderLogin(ctx, loginViewModel{
				form:      form,
				isInvalid: true,
			})
		}
		if err != nil {
			return err
		}

		err = sessions.Create(ctx.Context(), ctx.Res, session.User{
			Id:     u.ID,
			Roles:  []core.Role{u.Role},
			Claims: map[session.Claim]string{user.ClaimName: u.Name},
		})
		if err != nil {
			return err
		}
		if _, err := cs.Rotate(ctx.Res); err != nil { // the browser is somebody else now
			return err
		}
		return ctx.RedirectCode("/", http.StatusSeeOther)
	}
}

func handleLogout(sessions *session.Manager) owl.Handler {
	return func(ctx owl.Ctx) error {
		if err := sessions.Destroy(ctx.Res, ctx.Req); err != nil {
			return err
		}
		return ctx.RedirectCode(bootstrap.LoginURL, http.StatusSeeOther)
	}
}
