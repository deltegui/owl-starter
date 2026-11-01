package home

import "github.com/deltegui/owl-starter/sys/bootstrap"

func Bootstrap(ctx bootstrap.Context) {
	ctx.Mux.Get("/", handleShowIndex(), ctx.Auth())
	ctx.Mux.Post("/home/selectUsersAjax", handleUsersSelectAjax(), ctx.Auth())
	ctx.Mux.Post("/home/selectGenderAjax", handleGenderSelectAjax(), ctx.Auth())
}
