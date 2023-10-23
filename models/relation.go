package models

type Relation struct {
	ID         int64 `json:"id" gorm:"primaryKey"`                                             // 主键id
	UserID     int64 `json:"user_id" gorm:"index:idx_user_id;index:idx_user_follower"`         // 用户id
	FollowerID int64 `json:"follower_id" gorm:"index:idx_follower_id;index:idx_user_follower"` // 粉丝id
}
