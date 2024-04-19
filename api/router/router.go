package router

import (
	"context"
	"douyin/api/controller"
	"douyin/api/middleware"
	"douyin/config"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func Setup(conf *config.HertzConfig) *server.Hertz {
	h := server.Default(
		server.WithHostPorts(fmt.Sprintf("%s:%d", conf.Host, conf.Port)),
		server.WithMaxRequestBodySize(1024*1024*128),
	)

	// h.Use(middleware.RatelimitMiddleware(time.Millisecond, 100))

	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"message": "pong"})
	})

	apiRouter := h.Group("/douyin")

	// basic apis
	apiRouter.GET("/feed/", controller.Feed)

	userRouter := apiRouter.Group("/user")
	{
		userController := controller.NewUserController()
		userRouter.GET("/", userController.Info)
		userRouter.POST("/register/", userController.Register)
		userRouter.POST("/login/", userController.Login)
	}

	publishRouter := apiRouter.Group("/publish")
	{
		publishController := controller.NewPublishController()
		publishRouter.POST("/action/", middleware.AuthMiddleware(), publishController.Action)
		publishRouter.GET("/list/", publishController.List)
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
