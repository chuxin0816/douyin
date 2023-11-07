package dao

import (
	"context"
	"douyin/models"
	"douyin/pkg/snowflake"
	"douyin/response"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
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
		videoList = append(videoList, ToVideoResponse(userID, dVideo, authors[idx]))
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

	// 保存视频信息到数据库
	if err := db.Create(video).Error; err != nil {
		hlog.Error("mysql.SaveVideo: 保存视频信息到数据库失败")
		return err
	}

	// 修改用户发布视频数
	key := getRedisKey(KeyUserWorkCountPF + strconv.FormatInt(userID, 10))
	if err := rdb.Incr(context.Background(), key).Err(); err != nil {
		hlog.Error("redis.SaveVideo: 修改用户发布视频数失败")
		return err
	}

	return nil
}

// GetPublishList 获取用户发布的视频列表
func GetPublishList(userID, authorID int64) ([]*response.VideoResponse, error) {
	// 查询视频信息
	var dVideoList []*models.Video
	if err := db.Where("author_id = ?", authorID).Order("upload_time DESC").Find(&dVideoList).Error; err != nil {
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
	videoList := make([]*response.VideoResponse, 0, len(dVideoList))
	for _, dVideo := range dVideoList {
		videoList = append(videoList, ToVideoResponse(userID, dVideo, author))
	}

	return videoList, nil
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
		videoList = append(videoList, ToVideoResponse(userID, dVideo, authors[idx]))
	}

	return videoList, nil
}

func ToVideoResponse(userID int64, dVideo *models.Video, author *models.User) *response.VideoResponse {
	video := &response.VideoResponse{
		ID:            dVideo.ID,
		Author:        ToUserResponse(userID, author),
		CommentCount:  dVideo.CommentCount,
		PlayURL:       dVideo.PlayURL,
		CoverURL:      dVideo.CoverURL,
		FavoriteCount: dVideo.FavoriteCount,
		IsFavorite:    false,
		Title:         dVideo.Title,
	}
	// 未登录直接返回
	if userID == 0 {
		return video
	}

	// 查询缓存判断是否点赞
	key := getRedisKey(KeyUserFavoritePF + strconv.FormatInt(userID, 10))
	if rdb.SIsMember(context.Background(), key, dVideo.ID).Val() {
		video.IsFavorite = true
		return video
	}

	// 缓存未命中, 查询数据库
	// 获取用户喜欢列表
	videoIDs, err := GetFavoriteList(userID)
	if err != nil {
		hlog.Error("mysql.ToVideoResponse: 获取用户喜欢列表失败")
		return video
	}

	// 判断是否点赞
	for _, videoID := range videoIDs {
		if videoID == dVideo.ID {
			video.IsFavorite = true
			break
		}
	}

	// 将点赞信息写入缓存
	if video.IsFavorite {
		go func() {
			if err := rdb.SAdd(context.Background(), key, dVideo.ID).Err(); err != nil {
				hlog.Error("redis.ToVideoResponse: 将点赞信息写入缓存失败, err: ", err)
				return
			}
			if err := rdb.Expire(context.Background(), key, expireTime+randomDuration).Err(); err != nil {
				hlog.Error("redis.ToVideoResponse: 设置缓存过期时间失败, err: ", err)
				return
			}
		}()
	}
	return video
}
