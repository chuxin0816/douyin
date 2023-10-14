package mysql

import (
	"douyin/models"
	"douyin/response"
	"path"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

const (
	videoPrefix = "https://douyin-chuxin.oss-cn-shenzhen.aliyuncs.com/video/"
	imagePrefix = "https://douyin-chuxin.oss-cn-shenzhen.aliyuncs.com/image/"
)

func GetVideoList(latestTime time.Time, count int) (videoList []*response.VideoResponse, err error) {
	// 查询数据库
	year := latestTime.Year()
	if year < 1 || year > 9999 {
		hlog.Debug("mysql.GetVideoList: 参数时间超出限制")
		latestTime = time.Now()
	}
	var dVideoList []*models.Video
	err = db.Where("upload_time <= ?", latestTime).Order("upload_time DESC").Limit(count).Find(&dVideoList).Error
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
			ID:            dVideo.ID,
			Author:        authors[idx],
			CommentCount:  dVideo.CommentCount,
			PlayURL:       dVideo.PlayURL,
			CoverURL:      dVideo.CoverURL,
			FavoriteCount: dVideo.FavoriteCount,
			IsFavorite:    false, // 需要登录后通过用户id查询数据库判断
			Title:         dVideo.Title,
		})
	}
	return
}

func SaveVideo(userID int64, videoName, coverName, title string) error {
	// 保存视频信息到数据库
	video := &models.Video{
		AuthorID:   userID,
		PlayURL:    path.Join(videoPrefix, videoName),
		CoverURL:   path.Join(imagePrefix, coverName),
		UploadTime: time.Now(),
		Title:      title,
	}
	err := db.Create(video).Error
	if err != nil {
		hlog.Error("mysql.SaveVideo: 保存视频信息到数据库失败")
		return err
	}
	return nil
}
