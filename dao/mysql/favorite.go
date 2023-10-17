package mysql

import (
	"douyin/models"
	"errors"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

var (
	ErrAlreadyFavorite = errors.New("已经点赞过了")
)

func FavoriteAction(userID int64, videoID int64, actionType int) (err error) {
	// 查看是否已经点赞
	favorite := &models.Favorite{}
	err = db.Where("user_id = ? AND video_id = ?", userID, videoID).Find(favorite).Error
	if err != nil {
		hlog.Error("mysql.FavoriteAction: 查看是否已经点赞失败, err: ", err)
		return err
	}
	if favorite.ID != 0 {
		hlog.Error("mysql.FavoriteAction: 已经点赞过了")
		return ErrAlreadyFavorite
	}

	// 更新favorite表
	if actionType == 1 {
		err = db.Create(&models.Favorite{
			UserID:  userID,
			VideoID: videoID,
		}).Error
	} else {
		err = db.Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&models.Favorite{}).Error
	}
	if err != nil {
		hlog.Error("mysql.FavoriteAction: 更新favorite表失败, err: ", err)
		return err
	}

	// 更新video表中的favorite_count字段
	err = db.Model(&models.Video{}).Where("id = ?", videoID).Update("favorite_count", gorm.Expr("favorite_count + ?", actionType)).Error
	if err != nil {
		hlog.Error("mysql.FavoriteAction: 更新video表中的favorite_count字段失败, err: ", err)
		return err
	}

	// 更新user表中的favorite_count字段
	err = db.Model(models.User{}).Where("id = ?", userID).Update("favorite_count", gorm.Expr("favorite_count + ?", actionType)).Error
	if err != nil {
		hlog.Error("mysql.FavoriteAction: 更新user表中的favorite_count字段失败, err: ", err)
		return err
	}

	return nil
}
