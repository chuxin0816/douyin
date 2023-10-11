package mysql

import (
	"douyin/models"
	"douyin/response"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func GetVideoList(latestTime int64, count int) (videoList []*response.VideoResponse, err error) {
	// 查询数据库
	var dVideoList []*models.Video
	err = db.Where("upload_time <= ?", time.Unix(latestTime, 0)).Order("upload_time DESC").Limit(count).Find(&dVideoList).Error
	if err != nil {
		hlog.Error("mysql.GetVideoList: 查询数据库失败")
		return nil, err
	}

	// 通过作者id查询作者信息
	var authorIDs []string
	for _, dVideo := range dVideoList {
		authorIDs = append(authorIDs, strconv.FormatInt(dVideo.AuthorID, 10))
	}
	authors, err := GetUserByIDs(authorIDs)
	if err != nil {
		hlog.Error("mysql.GetVideoList: 通过作者id查询作者信息失败")
		return nil, err
	}

	// 将models.Video转换为response.VideoResponse
	for idx, dVideo := range dVideoList {
		videoList = append(videoList, &response.VideoResponse{
			Author:        authors[idx],
			CommentCount:  dVideo.CommentCount,
			CoverURL:      dVideo.CoverURL,
			FavoriteCount: dVideo.FavoriteCount,
			ID:            dVideo.ID,
			IsFavorite:    false, // 需要登录后通过用户id查询数据库判断
			PlayURL:       dVideo.PlayURL,
			Title:         dVideo.Title,
		})
	}
	return
}
