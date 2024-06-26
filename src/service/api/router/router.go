package router

import (
	"context"
	"fmt"
	"time"

	"douyin/src/config"
	"douyin/src/dal"
	"douyin/src/service/api/controller"
	"douyin/src/service/api/middleware"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/cache"
	"github.com/hertz-contrib/cache/persist"
	"github.com/hertz-contrib/http2/factory"
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
		tracer,
	)
	// HTTP2
	h.AddProtocol("h2", factory.NewServerFactory())

	// 链路追踪中间件
	h.Use(hertztracing.ServerMiddleware(cfg))

	// 缓存中间件
	memoryStore := persist.NewMemoryStore(10 * time.Second)
	h.Use(cache.NewCacheByRequestURI(
		memoryStore,
		2*time.Second,
		cache.WithPrefixKey(dal.Prefix),
	))

	// 限流中间件
	h.Use(middleware.RatelimitMiddleware(3000))

	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"message": "pong"})
	})

	apiRouter := h.Group("/douyin")

	// basic apis
	videoController := controller.NewVideoController()
	apiRouter.GET("/feed/", videoController.Feed)

	userRouter := apiRouter.Group("/user")
	{
		userController := controller.NewUserController()
		userRouter.GET("/", userController.Info)
		userRouter.POST("/register/", userController.Register)
		userRouter.POST("/login/", userController.Login)
	}

	publishRouter := apiRouter.Group("/publish")
	{
		publishRouter.POST("/action/", middleware.AuthMiddleware(), videoController.PublishAction)
		publishRouter.GET("/list/", videoController.PublishList)
	}

	// interaction apis
	favoriteRouter := apiRouter.Group("/favorite")
	{
		favoriteController := controller.NewFavoriteController()
		favoriteRouter.POST("/action/", middleware.AuthMiddleware(), favoriteController.Action)
		favoriteRouter.GET("/list/", favoriteController.List)
	}

	commentRouter := apiRouter.Group("/comment")
	{
		commentController := controller.NewCommentController()
		commentRouter.POST("/action/", middleware.AuthMiddleware(), commentController.Action)
		commentRouter.GET("/list/", commentController.List)
	}

	// social apis
	relationRouter := apiRouter.Group("/relation")
	{
		relationController := controller.NewRelationController()
		relationRouter.POST("/action/", middleware.AuthMiddleware(), relationController.Action)
		relationRouter.GET("/follow/list/", relationController.FollowList)
		relationRouter.GET("/follower/list/", relationController.FollowerList)
		relationRouter.GET("/friend/list/", relationController.FriendList)
	}

	messageRouter := apiRouter.Group("/message")
	{
		messageController := controller.NewMessageController()
		messageRouter.POST("/action/", middleware.AuthMiddleware(), messageController.Action)
		messageRouter.GET("/chat/", middleware.AuthMiddleware(), messageController.Chat)
	}

	return h
}
