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

type CommentController struct{}

func NewCommentController() *CommentController {
	return &CommentController{}
}

func (cc *CommentController) Action(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &models.CommentActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("CommentController.Action: 参数校验失败, err: ", err)
		return
	}

	// 验证token
	userID, err := jwt.ParseToken(req.Token)
	if err != nil {
		response.Error(ctx, response.CodeNoAuthority)
		hlog.Error("CommentController.Action: token无效, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := service.CommentAction(userID, req.ActionType, req.VideoID, req.CommentID, req.CommentText)
	if err != nil {
		if errors.Is(err, mysql.ErrCommentNotExist) {
			response.Error(ctx, response.CodeCommentNotExist)
			hlog.Error("controller.CommentAction: 评论不存在")
			return
		}
		response.Error(ctx, response.CodeServerBusy)
		hlog.Error("CommentController.Action: 业务逻辑处理失败, err: ", err)
		return
	}

	// 返回响应
	response.Success(ctx, resp)
}
