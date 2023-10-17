package models

type Favorite struct {
	ID      int64 `json:"id" gorm:"primaryKey"`                                        // 主键id
	UserID  int64 `json:"user_id" gorm:"index:idx_user_video;index:idx_user,unique"`   // 用户id
	VideoID int64 `json:"video_id" gorm:"index:idx_user_video;index:idx_video,unique"` // 视频id
}
