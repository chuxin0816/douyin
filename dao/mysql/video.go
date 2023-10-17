package mysql

import (
	"douyin/models"
	"douyin/response"
	"path"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

const (
	videoPrefix = "https://douyin-chuxin.oss-cn-shenzhen.aliyuncs.com/video/"
	imagePrefix = "https://douyin-chuxin.oss-cn-shenzhen.aliyuncs.com/image/"
)

// GetVideoList 获取视频Feed流
func GetVideoList(latestTime time.Time, count int) (videoList []*response.VideoResponse, nextTime *int64, err error) {
	// 查询数据库
	year := latestTime.Year()
	if year < 1 || year > 9999 {
		latestTime = time.Now()
	}
	var dVideoList []*models.Video
	err = db.Where("upload_time <= ?", latestTime).Order("upload_time DESC").Limit(count).Find(&dVideoList).Error
	if err != nil {
		hlog.Error("mysql.GetVideoList: 查询数据库失败")
		return nil, nil, err
	}

	// 通过作者id查询作者信息
	authorIDs := make([]int64, 0, len(dVideoList))
	for _, dVideo := range dVideoList {
		authorIDs = append(authorIDs, dVideo.AuthorID)
	}
	authors, err := GetUserByIDs(authorIDs)
	if err != nil {
		hlog.Error("mysql.GetVideoList: 通过作者id查询作者信息失败")
		return nil, nil, err
	}

	// 将models.Video转换为response.VideoResponse
	videoList = make([]*response.VideoResponse, 0, len(dVideoList))
	for idx, dVideo := range dVideoList {
		videoList = append(videoList, &response.VideoResponse{
			ID:            dVideo.ID,
			Author:        response.ToUserResponse(authors[idx]),
			CommentCount:  dVideo.CommentCount,
			PlayURL:       dVideo.PlayURL,
			CoverURL:      dVideo.CoverURL,
			FavoriteCount: dVideo.FavoriteCount,
			IsFavorite:    false, // 需要登录后通过用户id查询数据库判断
			Title:         dVideo.Title,
		})
	}
	if len(dVideoList) > 0 {
		nextTime = new(int64)
		*nextTime = dVideoList[len(dVideoList)-1].UploadTime.Unix()
	}
	return
}

// SaveVideo 保存视频信息到数据库
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

	// 修改用户发布视频数
	err = db.Model(&models.User{}).Where("id = ?", userID).Update("publish_count", gorm.Expr("publish_count + ?", 1)).Error
	if err != nil {
		hlog.Error("mysql.SaveVideo: 修改用户发布视频数失败")
		return err
	}
	return nil
}

// GetPublishList 获取用户发布的视频列表
func GetPublishList(authorID int64) (videoList []*response.VideoResponse, err error) {
	// 查询数据库
	var dVideoList []*models.Video
	err = db.Where("author_id = ?", authorID).Order("upload_time DESC").Find(&dVideoList).Error
	if err != nil {
		hlog.Error("mysql.GetPublishList: 查询数据库失败")
		return nil, err
	}

	// 查询作者信息
	author, err := GetUserByID(authorID)
	if err != nil {
		hlog.Error("mysql.GetPublishList: 查询作者信息失败")
		return nil, err
	}

	// 将models.Video转换为response.VideoResponse
	videoList = make([]*response.VideoResponse, 0, len(dVideoList))
	for _, dVideo := range dVideoList {
		videoList = append(videoList, &response.VideoResponse{
			ID:            dVideo.ID,
			Author:        response.ToUserResponse(author),
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
