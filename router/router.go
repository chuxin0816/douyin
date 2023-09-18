package router

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/spf13/viper"
)

func Setup() *server.Hertz {
	h := server.Default(server.WithHostPorts(
		fmt.Sprintf("%s:%d", viper.GetString("hertz.host"), viper.GetInt("hertz.port")),
	))

	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"message": "pong"})
	})

	return h
}
