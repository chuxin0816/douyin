package mysql

import (
	"douyin/models"
	"douyin/pkg/snowflake"
	"douyin/response"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

const (
	videoPrefix = "http://oss.chuxin0816.com/video/"
	imagePrefix = "http://oss.chuxin0816.com/image/"
)

// GetVideoList 获取视频Feed流
func GetFeedList(userID int64, latestTime time.Time, count int) (videoList []*response.VideoResponse, nextTime *int64, err error) {
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
		return nil, nil, err
	}

	// 将models.Video转换为response.VideoResponse
	videoList = make([]*response.VideoResponse, 0, len(dVideoList))
	for idx, dVideo := range dVideoList {
		videoList = append(videoList, &response.VideoResponse{
			ID:            dVideo.ID,
			Author:        ToUserResponse(userID, authors[idx]),
			CommentCount:  dVideo.CommentCount,
			PlayURL:       dVideo.PlayURL,
			CoverURL:      dVideo.CoverURL,
			FavoriteCount: dVideo.FavoriteCount,
			IsFavorite:    false,
			Title:         dVideo.Title,
		})
	}

	// 通过用户id查询是否点赞
	if userID > 0 && len(dVideoList) > 0 {
		// 获取用户喜欢列表
		videoIDs, err := GetFavoriteList(userID)
		if err != nil {
			return nil, nil, err
		}

		// 将喜欢列表转换为map加快查询速度
		videoIDMap := make(map[int64]struct{}, len(videoIDs))
		for _, v := range videoIDs {
			videoIDMap[v] = struct{}{}
		}

		// 判断是否点赞
		for _, video := range videoList {
			if _, exist := videoIDMap[video.ID]; exist {
				video.IsFavorite = true
			}
		}
	}

	// 计算下次请求的时间
	if len(dVideoList) > 0 {
		nextTime = new(int64)
		*nextTime = dVideoList[len(dVideoList)-1].UploadTime.Unix()
	}
	return
}

// SaveVideo 保存视频信息到数据库
func SaveVideo(userID int64, videoName, coverName, title string) error {
	video := &models.Video{
		ID:         snowflake.GenerateID(),
		AuthorID:   userID,
		PlayURL:    videoPrefix + videoName,
		CoverURL:   imagePrefix + coverName,
		UploadTime: time.Now(),
		Title:      title,
	}

	// 开启事务
	err := db.Transaction(func(tx *gorm.DB) error {
		// 保存视频信息到数据库
		if err := db.Create(video).Error; err != nil {
			hlog.Error("mysql.SaveVideo: 保存视频信息到数据库失败")
			return err
		}

		// 修改用户发布视频数
		if err := db.Model(&models.User{}).Where("id = ?", userID).Update("work_count", gorm.Expr("work_count + ?", 1)).Error; err != nil {
			hlog.Error("mysql.SaveVideo: 修改用户发布视频数失败")
			return err
		}

		// 提交事务
		return nil
	})

	return err
}

// GetPublishList 获取用户发布的视频列表
func GetPublishList(userID, authorID int64) (videoList []*response.VideoResponse, err error) {
	// 查询视频信息
	var dVideoList []*models.Video
	err = db.Where("author_id = ?", authorID).Order("upload_time DESC").Find(&dVideoList).Error
	if err != nil {
		hlog.Error("mysql.GetPublishList: 查询视频信息失败")
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
			Author:        ToUserResponse(userID, author),
			CommentCount:  dVideo.CommentCount,
			PlayURL:       dVideo.PlayURL,
			CoverURL:      dVideo.CoverURL,
			FavoriteCount: dVideo.FavoriteCount,
			IsFavorite:    false,
			Title:         dVideo.Title,
		})
	}

	// 通过用户id查询是否点赞
	if len(videoList) > 0 {
		// 获取用户喜欢列表
		videoIDs, err := GetFavoriteList(userID)
		if err != nil {
			return nil, err
		}

		// 将喜欢列表转换为map加快查询速度
		videoIDMap := make(map[int64]struct{}, len(videoIDs))
		for _, v := range videoIDs {
			videoIDMap[v] = struct{}{}
		}

		// 判断是否点赞
		for _, video := range videoList {
			if _, exist := videoIDMap[video.ID]; exist {
				video.IsFavorite = true
			}
		}
	}
	return
}

func GetVideoList(userID int64, videoIDs []int64) ([]*response.VideoResponse, error) {
	// 查询视频信息
	var dVideoList []*models.Video
	err := db.Where("id IN ?", videoIDs).Order("upload_time DESC").Find(&dVideoList).Error
	if err != nil {
		hlog.Error("mysql.GetVideoList: 查询视频信息失败")
		return nil, err
	}

	// 通过作者id查询作者信息
	authorIDs := make([]int64, 0, len(dVideoList))
	for _, dVideo := range dVideoList {
		authorIDs = append(authorIDs, dVideo.AuthorID)
	}
	authors, err := GetUserByIDs(authorIDs)
	if err != nil {
		return nil, err
	}

	// 将models.Video转换为response.VideoResponse
	videoList := make([]*response.VideoResponse, 0, len(dVideoList))
	for idx, dVideo := range dVideoList {
		videoList = append(videoList, &response.VideoResponse{
			ID:            dVideo.ID,
			Author:        ToUserResponse(userID, authors[idx]),
			CommentCount:  dVideo.CommentCount,
			PlayURL:       dVideo.PlayURL,
			CoverURL:      dVideo.CoverURL,
			FavoriteCount: dVideo.FavoriteCount,
			IsFavorite:    false,
			Title:         dVideo.Title,
		})
	}

	// 通过用户id查询是否点赞
	if userID > 0 && len(dVideoList) > 0 {
		// 获取用户喜欢列表
		videoIDs, err := GetFavoriteList(userID)
		if err != nil {
			return nil, err
		}

		// 将喜欢列表转换为map加快查询速度
		videoIDMap := make(map[int64]struct{}, len(videoIDs))
		for _, v := range videoIDs {
			videoIDMap[v] = struct{}{}
		}

		// 判断是否点赞
		for _, video := range videoList {
			if _, exist := videoIDMap[video.ID]; exist {
				video.IsFavorite = true
			}
		}
	}

	return videoList, nil
}
