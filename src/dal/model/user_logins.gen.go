// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameUserLogin = "user_logins"

// UserLogin mapped from table <user_logins>
type UserLogin struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	Username   string    `gorm:"column:username;not null;comment:用户名" json:"username"`                                  // 用户名
	Password   string    `gorm:"column:password;not null;comment:加密密码" json:"password"`                                 // 加密密码
	CreateTime time.Time `gorm:"column:create_time;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"create_time"` // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"update_time"` // 更新时间
}

// TableName UserLogin's table name
func (*UserLogin) TableName() string {
	return TableNameUserLogin
}
