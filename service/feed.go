package service

import (
	"douyin/dao/mysql"
	"douyin/models"
	"douyin/response"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

const count = 30

func Feed(req *models.FeedRequest) (*response.FeedResponse, error) {
	// 解析请求
	if len(req.LatestTime) == 0 {
		req.LatestTime = "0"
	}
	latestTime, err := strconv.ParseInt(req.LatestTime, 10, 64)
	if err != nil {
		hlog.Error("service.Feed: 解析请求失败")
		return nil, err
	}

	// 查询视频列表
	videoList, err := mysql.GetVideoList(latestTime, count)
	if err != nil {
		hlog.Error("service.Feed: 查询视频列表失败")
		return nil, err
	}
	return &response.FeedResponse{
		Response:  &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		NextTime:  time.Now().Unix(),
		VideoList: videoList,
	}, nil
}
