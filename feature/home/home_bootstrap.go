package home

import "github.com/deltegui/owl-starter/sys/bootstrap"

func Bootstrap(ctx bootstrap.Context) {
	ctx.Mux.Get("/", handleShowIndex(), ctx.Auth())
}
