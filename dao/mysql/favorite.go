package mysql

import (
	"douyin/models"
	"errors"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

var (
	ErrAlreadyFavorite = errors.New("已经点赞过了")
	ErrNotFavorite     = errors.New("还没有点赞过")
)

func FavoriteAction(userID int64, videoID int64, actionType int64) (err error) {
	// 查看是否已经点赞
	favorite := &models.Favorite{}
	err = db.Where("user_id = ? AND video_id = ?", userID, videoID).Find(favorite).Error
	if err != nil {
		hlog.Error("mysql.FavoriteAction: 查看是否已经点赞失败, err: ", err)
		return err
	}
	if favorite.ID != 0 && actionType == 1 {
		hlog.Error("mysql.FavoriteAction: 已经点赞过了")
		return ErrAlreadyFavorite
	} else if favorite.ID == 0 && actionType == -1 {
		hlog.Error("mysql.FavoriteAction: 还没有点赞过")
		return ErrNotFavorite
	}

	// 开启事务
	tx := db.Begin()

	// 更新favorite表
	if actionType == 1 {
		err = tx.Create(&models.Favorite{
			UserID:  userID,
			VideoID: videoID,
		}).Error
	} else {
		err = tx.Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&models.Favorite{}).Error
	}
	if err != nil {
		tx.Rollback()
		hlog.Error("mysql.FavoriteAction: 更新favorite表失败, err: ", err)
		return err
	}

	// 更新video表中的favorite_count字段
	err = tx.Model(&models.Video{}).Where("id = ?", videoID).Update("favorite_count", gorm.Expr("favorite_count + ?", actionType)).Error
	if err != nil {
		tx.Rollback()
		hlog.Error("mysql.FavoriteAction: 更新video表中的favorite_count字段失败, err: ", err)
		return err
	}

	// 更新user表中的favorite_count字段
	err = tx.Model(models.User{}).Where("id = ?", userID).Update("favorite_count", gorm.Expr("favorite_count + ?", actionType)).Error
	if err != nil {
		tx.Rollback()
		hlog.Error("mysql.FavoriteAction: 更新user表中的favorite_count字段失败, err: ", err)
		return err
	}

	// 更新user表中的total_favorited字段
	err = tx.Model(models.User{}).Where("id = ?", userID).Update("total_favorited", gorm.Expr("total_favorited + ?", actionType)).Error
	if err != nil {
		tx.Rollback()
		hlog.Error("mysql.FavoriteAction: 更新user表中的total_favorited字段失败, err: ", err)
		return err
	}

	// 提交事务
	tx.Commit()

	return nil
}

func GetFavoriteList(userID int64) ([]int64, error) {
	var favoriteList []*models.Favorite
	err := db.Where("user_id = ?", userID).Find(&favoriteList).Error
	if err != nil {
		hlog.Error("mysql.GetFavoriteList: 查询favorite表失败, err: ", err)
		return nil, err
	}

	// 将models.Favorite视频ID提取出来
	videoIDs := make([]int64, 0, len(favoriteList))
	for _, favorite := range favoriteList {
		videoIDs = append(videoIDs, favorite.VideoID)
	}
	return videoIDs, nil
}
