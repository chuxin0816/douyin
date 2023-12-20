// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

const TableNameUser = "users"

// User mapped from table <users>
type User struct {
	ID              int64  `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Name            string `gorm:"column:name" json:"name"`
	Avatar          string `gorm:"column:avatar" json:"avatar"`
	BackgroundImage string `gorm:"column:background_image" json:"background_image"`
	TotalFavorited  string `gorm:"column:total_favorited;default:0" json:"total_favorited"`
	FavoriteCount   int64  `gorm:"column:favorite_count" json:"favorite_count"`
	FollowCount     int64  `gorm:"column:follow_count" json:"follow_count"`
	FollowerCount   int64  `gorm:"column:follower_count" json:"follower_count"`
	WorkCount       int64  `gorm:"column:work_count" json:"work_count"`
	Password        string `gorm:"column:password" json:"password"`
	Signature       string `gorm:"column:signature" json:"signature"`
}

// TableName User's table name
func (*User) TableName() string {
	return TableNameUser
}
