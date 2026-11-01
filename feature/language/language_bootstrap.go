package language

import "github.com/deltegui/owl-starter/sys/bootstrap"

func Bootstrap(ctx bootstrap.Context) {
	ctx.Mux.Post("/language/:lang", handleChange())
}
