package router

import (
	"context"
	"douyin/config"
	"douyin/controller"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func Setup(conf *config.HertzConfig) *server.Hertz {
	h := server.Default(server.WithHostPorts(
		fmt.Sprintf("%s:%d", conf.Host, conf.Port),
	))

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
		publishRouter.POST("/action/", publishController.Action)
		// publishRouter.GET("/list/", publishController.PublishList)
	}

	// interaction apis
	// apiRouter.POST("/favorite/action", controller.FavoriteAction)
	// apiRouter.GET("/favorite/list", controller.FavoriteList)
	// apiRouter.POST("/comment/action", controller.CommentAction)
	// apiRouter.GET("/comment/list", controller.CommentList)

	// social apis
	// apiRouter.POST("/relation/action", controller.RelationAction)
	// apiRouter.GET("/relation/follow/list", controller.FollowList)
	// apiRouter.GET("/relation/follower/list", controller.FollowerList)
	// apiRouter.GET("/relation/friend/list", controller.FriendList)
	// apiRouter.GET("/message/chat", controller.MessageChat)
	// apiRouter.POST("/message/action", controller.MessageAction)

	return h
}
