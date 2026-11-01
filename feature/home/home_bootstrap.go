package home

import "github.com/deltegui/owl-starter/sys/bootstrap"

func Bootstrap(ctx bootstrap.Context) {
	ctx.Mux.Get("/", handleShowIndex(), ctx.Auth())
	ctx.Mux.Get("/home/selectUsersAjax", handleUsersSelectAjax(), ctx.Auth())
	ctx.Mux.Get("/home/selectGenderAjax", handleGenderSelectAjax(), ctx.Auth())
	ctx.Mux.Get("/home/report", handleReport(), ctx.Auth())
	ctx.Mux.Get("/home/people", handlePeople(), ctx.Auth())
}
