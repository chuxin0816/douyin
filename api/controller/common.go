package controller

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

const (
	CtxUserIDKey = "userID"
	rpcErrPrefix = "remote or network error[remote]: biz error: "
)

const (
	CodeSuccess     int32 = 0
	CodeNoAuthority int32 = 1000 + iota
	CodeInvalidParam
	CodeUserExist
	CodeUserNotExist
	CodeInvalidPassword
	CodeFileTooLarge
	CodeAlreadyFavorite
	CodeNotFavorite
	CodeVideoNotExist
	CodeCommentNotExist
	CodeAlreadyFollow
	CodeNotFollow
	CodeFollowLimit
	CodeServerBusy
)

var codeMsgMap = map[int32]string{
	CodeSuccess:         "请求成功",
	CodeNoAuthority:     "权限不足",
	CodeInvalidParam:    "请求参数错误",
	CodeUserExist:       "用户名已存在",
	CodeUserNotExist:    "用户名不存在",
	CodeInvalidPassword: "密码错误",
	CodeFileTooLarge:    "文件太大",
	CodeAlreadyFavorite: "已经点赞过了",
	CodeNotFavorite:     "还没有点赞过",
	CodeVideoNotExist:   "视频不存在",
	CodeCommentNotExist: "评论不存在",
	CodeAlreadyFollow:   "已经关注过了",
	CodeNotFollow:       "还没有关注过",
	CodeFollowLimit:     "关注数超过限制",
	CodeServerBusy:      "服务器繁忙",
}

type Response struct {
	StatusCode int32  `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  string `json:"status_msg"`  // 返回状态描述
}

func Success(ctx *app.RequestContext, data any) {
	ctx.JSON(consts.StatusOK, data)
}

func Error(ctx *app.RequestContext, code int32) {
	ctx.JSON(consts.StatusOK, &Response{StatusCode: code, StatusMsg: StatusMsg(code)})
}

func StatusMsg(code int32) string {
	msg, ok := codeMsgMap[code]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}

func errorIs(err error, target error) bool {
	return err.Error() == rpcErrPrefix+target.Error()
}
