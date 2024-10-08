package router

import (
	"context"
	"fmt"

	"douyin/src/common/mtl"
	"douyin/src/config"
	"douyin/src/service/api/controller"
	"douyin/src/service/api/mw"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/cors"
	"github.com/hertz-contrib/gzip"
	"github.com/hertz-contrib/http2/factory"
	hertzprom "github.com/hertz-contrib/monitor-prometheus"
	hertztracing "github.com/hertz-contrib/obs-opentelemetry/tracing"
)

func Setup(conf *config.HertzConfig) *server.Hertz {
	tracer, cfg := hertztracing.NewServerTracer()
	h := server.Default(
		server.WithHostPorts(fmt.Sprintf("%s:%d", conf.Host, conf.Port)),
		server.WithMaxRequestBodySize(512*1024*1024),
		server.WithStreamBody(true),
		server.WithH2C(true),
		server.WithALPN(true),
		server.WithTracer(hertzprom.NewServerTracer(
			"",
			"",
			hertzprom.WithRegistry(mtl.Registry),
			hertzprom.WithDisableServer(true),
		)),
		tracer,
	)
	h.Use(hertztracing.ServerMiddleware(cfg))

	registerMiddleware(h)

	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"message": "pong"})
	})

	apiRouter := h.Group("/douyin")

	// basic apis
	videoController := controller.NewVideoController()
	apiRouter.GET("/feed/", videoController.Feed)
	apiRouter.GET("/refresh_token", mw.RefreshTokenMiddleware())

	userRouter := apiRouter.Group("/user")
	{
		userController := controller.NewUserController()
		userRouter.GET("/", userController.Info)
		userRouter.POST("/register/", userController.Register)
		userRouter.POST("/login/", userController.Login)
	}

	publishRouter := apiRouter.Group("/publish")
	{
		publishRouter.POST("/action/", mw.AuthMiddleware(), videoController.PublishAction)
		publishRouter.GET("/list/", videoController.PublishList)
	}

	// interaction apis
	favoriteRouter := apiRouter.Group("/favorite")
	{
		favoriteController := controller.NewFavoriteController()
		favoriteRouter.POST("/action/", mw.AuthMiddleware(), favoriteController.Action)
		favoriteRouter.GET("/list/", favoriteController.List)
	}

	commentRouter := apiRouter.Group("/comment")
	{
		commentController := controller.NewCommentController()
		commentRouter.POST("/action/", mw.AuthMiddleware(), commentController.Action)
		commentRouter.GET("/list/", commentController.List)
	}

	// social apis
	relationRouter := apiRouter.Group("/relation")
	{
		relationController := controller.NewRelationController()
		relationRouter.POST("/action/", mw.AuthMiddleware(), relationController.Action)
		relationRouter.GET("/follow/list/", relationController.FollowList)
		relationRouter.GET("/follower/list/", relationController.FollowerList)
		relationRouter.GET("/friend/list/", relationController.FriendList)
	}

	messageRouter := apiRouter.Group("/message")
	{
		messageController := controller.NewMessageController()
		messageRouter.POST("/action/", mw.AuthMiddleware(), messageController.Action)
		messageRouter.GET("/chat/", mw.AuthMiddleware(), messageController.Chat)
	}

	return h
}

func registerMiddleware(h *server.Hertz) {
	// HTTP2
	h.AddProtocol("h2", factory.NewServerFactory())

	// cores
	h.Use(cors.Default())

	// gzip
	h.Use(gzip.Gzip(gzip.DefaultCompression))

	// 限流中间件
	h.Use(mw.RatelimitMiddleware(3000, 2000))
}
