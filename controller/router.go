package controller

import (
	"context"
	"douyin/config"
	"douyin/middleware"
	"fmt"
	"time"

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

	h.Use(middleware.RatelimitMiddleware(time.Millisecond, 100))

	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"message": "pong"})
	})

	apiRouter := h.Group("/douyin")

	// basic apis
	apiRouter.GET("/feed/", Feed)

	userRouter := apiRouter.Group("/user")
	{
		userController := NewUserController()
		userRouter.GET("/", middleware.AuthMiddleware(), userController.Info)
		userRouter.POST("/register/", userController.Register)
		userRouter.POST("/login/", userController.Login)
	}

	publishRouter := apiRouter.Group("/publish", middleware.AuthMiddleware())
	{
		publishController := NewPublishController()
		publishRouter.POST("/action/", publishController.Action)
		publishRouter.GET("/list/", publishController.List)
	}

	// interaction apis
	favoriteRouter := apiRouter.Group("/favorite", middleware.AuthMiddleware())
	{
		favoriteController := NewFavoriteController()
		favoriteRouter.POST("/action/", favoriteController.Action)
		favoriteRouter.GET("/list/", favoriteController.List)
	}

	commentRouter := apiRouter.Group("/comment", middleware.AuthMiddleware())
	{
		commentController := NewCommentController()
		commentRouter.POST("/action/", commentController.Action)
		commentRouter.GET("/list/", commentController.List)
	}

	// social apis
	relationRouter := apiRouter.Group("/relation", middleware.AuthMiddleware())
	{
		relationController := NewRelationController()
		relationRouter.POST("/action/", relationController.Action)
		relationRouter.GET("/follow/list/", relationController.FollowList)
		relationRouter.GET("/follower/list/", relationController.FollowerList)
		relationRouter.GET("/friend/list/", relationController.FriendList)
	}

	messageRouter := apiRouter.Group("/message", middleware.AuthMiddleware())
	{
		messageController := NewMessageController()
		messageRouter.POST("/action/", messageController.Action)
		messageRouter.GET("/chat/", messageController.Chat)
	}

	return h
}
