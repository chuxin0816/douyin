package service

import (
	"douyin/dao/mysql"
	"douyin/models"
	"douyin/response"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

const count = 30

func Feed(req *models.FeedRequest, userID int64) (resp *response.FeedResponse, err error) {
	// 解析请求
	var latestTime int64
	latestTime = req.LatestTime
	if latestTime == 0 {
		latestTime = time.Now().Unix()
	}

	// 查询视频列表
	videoList, nextTime, err := mysql.GetVideoList(time.Unix(latestTime, 0), count)
	if err != nil {
		hlog.Error("service.Feed: 查询视频列表失败")
		return nil, err
	}

	// TODO: 通过用户id查询数据库判断是否点赞

	// 返回响应
	return &response.FeedResponse{
		Response:  &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		VideoList: videoList,
		NextTime:  nextTime,
	}, nil
}
