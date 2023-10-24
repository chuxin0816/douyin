package mysql

import (
	"douyin/models"
	"errors"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

var (
	ErrAlreadyFollow = errors.New("已经关注过了")
	ErrNotFollow     = errors.New("还没有关注过")
)

func RelationAction(userID, toUserID int64, actionType int) error {
	// 查看是否关注
	relation := &models.Relation{}
	err := db.Where("user_id = ? AND follower_id = ?", toUserID, userID).Find(relation).Error
	if err != nil {
		hlog.Error("mysql.RelationAction 查看是否关注失败, err: ", err)
		return err
	}
	if relation.ID != 0 && actionType == 1 {
		hlog.Error("mysql.RelationAction 已经关注过了")
		return ErrAlreadyFollow
	} else if relation.ID == 0 && actionType == -1 {
		hlog.Error("mysql.RelationAction 还没有关注过")
		return ErrNotFollow
	}

	// 开启事务
	tx := db.Begin()

	// 更新relation表
	if actionType == 1 {
		err = tx.Create(&models.Relation{
			UserID:     toUserID,
			FollowerID: userID,
		}).Error
	} else {
		err = tx.Where("user_id = ? AND follower_id = ?", toUserID, userID).Delete(&models.Relation{}).Error
	}
	if err != nil {
		tx.Rollback()
		hlog.Error("mysql.RelationAction 更新relation表失败, err: ", err)
		return err
	}

	// 更新user表
	err = tx.Model(&models.User{}).Where("id = ?", userID).Update("follow_count", gorm.Expr("follow_count + ?", actionType)).Error
	if err != nil {
		tx.Rollback()
		hlog.Error("mysql.RelationAction 更新user表follow_count字段失败, err: ", err)
		return err
	}
	err = tx.Model(&models.User{}).Where("id = ?", toUserID).Update("follower_count", gorm.Expr("follower_count + ?", actionType)).Error
	if err != nil {
		tx.Rollback()
		hlog.Error("mysql.RelationAction 更新user表follower_count字段失败, err: ", err)
		return err
	}

	// 提交事务
	tx.Commit()

	return nil
}

func FollowList(userID, toUserID int64) ([]*models.User, error) {
	// 查询用户ID列表
	var userIDList []int64
	err := db.Table("relations").Select("user_id").Where("follower_id = ?", toUserID).Find(&userIDList).Error
	if err != nil {
		hlog.Error("mysql.FollowList 查询用户ID列表失败, err: ", err)
		return nil, err
	}

	// 查询用户列表
	userList, err := GetUserByIDs(userID, userIDList)
	if err != nil {
		hlog.Error("mysql.FollowList 查询用户列表失败, err: ", err)
		return nil, err
	}

	return userList, nil
}

func FollowerList(userID, toUserID int64) ([]*models.User, error) {
	// 查询粉丝ID列表
	var followerIDList []int64
	err := db.Table("relations").Select("follower_id").Where("user_id = ?", toUserID).Find(&followerIDList).Error
	if err != nil {
		hlog.Error("mysql.FollowerList 查询粉丝ID列表失败, err: ", err)
		return nil, err
	}

	// 查询粉丝列表
	followerList, err := GetUserByIDs(userID, followerIDList)
	if err != nil {
		hlog.Error("mysql.FollowerList 查询粉丝列表失败, err: ", err)
		return nil, err
	}

	return followerList, nil
}
