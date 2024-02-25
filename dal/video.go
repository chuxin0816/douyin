package dal

import (
	"context"
	"douyin/dal/model"
	"douyin/pkg/snowflake"
	"douyin/rpc/kitex_gen/feed"
	"strconv"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

const (
	videoPrefix = "http://oss.chuxin0816.com/video/"
	imagePrefix = "http://oss.chuxin0816.com/image/"
)

// GetVideoList 获取视频Feed流
func GetFeedList(userID *int64, latestTime time.Time, count int) (videoList []*feed.Video, nextTime *int64, err error) {
	// 查询数据库
	mVideoList, err := qVideo.WithContext(context.Background()).Where(qVideo.UploadTime.Lte(latestTime)).
		Order(qVideo.UploadTime.Desc()).Limit(count).Find()
	if err != nil {
		klog.Error("查询数据库失败")
		return nil, nil, err
	}

	// 通过作者id查询作者信息
	authorIDs := make([]int64, len(mVideoList))
	for i, mVideo := range mVideoList {
		authorIDs[i] = mVideo.AuthorID
	}
	authors, err := GetUserByIDs(authorIDs)
	if err != nil {
		return nil, nil, err
	}

	// 将model.Video转换为response.VideoResponse
	videoList = make([]*feed.Video, len(mVideoList))
	for idx, mVideo := range mVideoList {
		videoList = append(videoList, ToVideoResponse(userID, mVideo, authors[idx]))
	}

	// 计算下次请求的时间
	if len(mVideoList) > 0 {
		nextTime = new(int64)
		*nextTime = mVideoList[len(mVideoList)-1].UploadTime.Unix()
	}
	return
}

// SaveVideo 保存视频信息到数据库
func SaveVideo(userID int64, videoName, coverName, title string) error {
	video := &model.Video{
		ID:         snowflake.GenerateID(),
		AuthorID:   userID,
		PlayURL:    videoPrefix + videoName,
		CoverURL:   imagePrefix + coverName,
		UploadTime: time.Now(),
		Title:      title,
	}

	// 保存视频信息到数据库
	if err := qVideo.WithContext(context.Background()).Create(video); err != nil {
		klog.Error("保存视频信息到数据库失败")
		return err
	}

	// 修改用户发布视频数
	key := GetRedisKey(KeyUserWorkCountPF + strconv.FormatInt(userID, 10))
	if err := RDB.Incr(context.Background(), key).Err(); err != nil {
		klog.Error("修改用户发布视频数失败")
		return err
	}

	// 添加到布隆过滤器
	bloomFilter.Add([]byte(strconv.FormatInt(video.ID, 10)))

	// 写入待同步队列
	CacheUserID.Store(userID, struct{}{})

	return nil
}

// GetPublishList 获取用户发布的视频列表
func GetPublishList(userID *int64, authorID int64) ([]*feed.Video, error) {
	// 查询视频信息
	mVideoList, err := qVideo.WithContext(context.Background()).Where(qVideo.AuthorID.Eq(authorID)).
		Order(qVideo.UploadTime.Desc()).Find()
	if err != nil {
		klog.Error("查询视频信息失败")
		return nil, err
	}

	// 查询作者信息
	author, err := GetUserByID(authorID)
	if err != nil {
		klog.Error("查询作者信息失败")
		return nil, err
	}

	// 将model.Video转换为response.VideoResponse
	videoList := make([]*feed.Video, 0, len(mVideoList))
	for _, mVideo := range mVideoList {
		videoList = append(videoList, ToVideoResponse(userID, mVideo, author))
	}

	return videoList, nil
}

func GetVideoList(userID *int64, videoIDs []int64) ([]*feed.Video, error) {
	// 查询视频信息
	mVideoList, err := qVideo.WithContext(context.Background()).Where(qVideo.ID.In(videoIDs...)).
		Order(qVideo.UploadTime.Desc()).Find()
	if err != nil {
		klog.Error("查询视频信息失败")
		return nil, err
	}

	// 通过作者id查询作者信息
	authorIDs := make([]int64, len(mVideoList))
	for i, mVideo := range mVideoList {
		authorIDs[i] = mVideo.AuthorID
	}
	authors, err := GetUserByIDs(authorIDs)
	if err != nil {
		return nil, err
	}

	// 将model.Video转换为feed.Video
	videoList := make([]*feed.Video, len(mVideoList))
	for i, mVideo := range mVideoList {
		videoList[i] = ToVideoResponse(userID, mVideo, authors[i])
	}

	return videoList, nil
}

func ToVideoResponse(userID *int64, mVideo *model.Video, author *model.User) *feed.Video {
	video := &feed.Video{
		Id:            mVideo.ID,
		Author:        ToUserResponse(userID, author),
		CommentCount:  mVideo.CommentCount,
		PlayUrl:       mVideo.PlayURL,
		CoverUrl:      mVideo.CoverURL,
		FavoriteCount: mVideo.FavoriteCount,
		IsFavorite:    false,
		Title:         mVideo.Title,
	}
	// 未登录直接返回
	if userID == nil || *userID == 0 {
		return video
	}

	// 使用singleflight避免缓存击穿和减少缓存压力
	// 查询缓存判断是否点赞
	key := GetRedisKey(KeyUserFavoritePF + strconv.FormatInt(*userID, 10))
	g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()
		if RDB.SIsMember(context.Background(), key, mVideo.ID).Val() {
			video.IsFavorite = true
			return nil, nil
		}

		// 缓存未命中, 查询数据库
		favorite, err := qFavorite.WithContext(context.Background()).Where(qFavorite.UserID.Eq(*userID), qFavorite.VideoID.Eq(mVideo.ID)).
			Select(qFavorite.ID).First()
		if err != nil {
			klog.Error("查询favorite表失败, err: ", err)
			return nil, err
		}
		if favorite.ID != 0 {
			video.IsFavorite = true
			// 写入缓存
			go func() {
				if err := RDB.SAdd(context.Background(), key, mVideo.ID).Err(); err != nil {
					klog.Error("将点赞信息写入缓存失败, err: ", err)
					return
				}
				if err := RDB.Expire(context.Background(), key, ExpireTime+GetRandomTime()).Err(); err != nil {
					klog.Error("设置缓存过期时间失败, err: ", err)
					return
				}
			}()
		}
		return nil, nil
	})

	return video
}

func GetAuthorID(videoID int64) (int64, error) {
	// 先查询作者的ID
	var authorID int64
	err := qVideo.WithContext(context.Background()).Where(qVideo.ID.Eq(videoID)).Select(qVideo.AuthorID).Scan(&authorID)

	return authorID, err
}

func UpdateVideo(video *model.Video) error {
	_, err := qVideo.WithContext(context.Background()).Where(qVideo.ID.Eq(video.ID)).Updates(map[string]any{
		"favorite_count": video.FavoriteCount,
		"comment_count":  video.CommentCount,
	})
	return err
}
