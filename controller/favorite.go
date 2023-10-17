package controller

import (
	"context"
	"douyin/dao/mysql"
	"douyin/models"
	"douyin/pkg/jwt"
	"douyin/response"
	"douyin/service"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type FavoriteController struct{}

func NewFavoriteController() *FavoriteController {
	return &FavoriteController{}
}

func (fc *FavoriteController) Action(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &models.FavoriteActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("FavoriteController: 参数校验失败, err: ", err)
		return
	}

	// 验证token
	userID, err := jwt.ParseToken(req.Token)
	if err != nil {
		if errors.Is(err, jwt.ErrInvalidToken) {
			response.Error(ctx, response.CodeNoAuthority)
			hlog.Error("FavoriteController.Action: token无效")
			return
		}
		response.Error(ctx, response.CodeServerBusy)
		hlog.Error("FavoriteController.Action: jwt解析出错, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := service.FavoriteAction(userID, req.VideoID, req.ActionType)
	if err != nil {
		if errors.Is(err, mysql.ErrAlreadyFavorite) {
			response.Error(ctx, response.CodeAlreadyFavorite)
			hlog.Error("FavoriteController.Action: 已经点赞过了")
			return
		}
		if errors.Is(err, mysql.ErrNotFavorite) {
			response.Error(ctx, response.CodeNotFavorite)
			hlog.Error("FavoriteController.Action: 还没有点赞过")
			return
		}
		response.Error(ctx, response.CodeServerBusy)
		hlog.Error("FavoriteController.Action: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	response.Success(ctx, resp)
}
