package models

import "time"

type User struct {
	ID         int       `json:"id" gorm:"primaryKey;column:id"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	Phone      string    `json:"phone"`
	Gender     int       `json:"gender"`
	UserStatus int       `json:"user_status"`
	Email      string    `json:"email"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
	Role       int       `json:"role"`
	Avatar     string    `json:"avatar"`
	IsDelete   int       `json:"is_delete"`
}

// TableName 设置表名为 sys_user
func (User) TableName() string {
	return "sys_user"
}
