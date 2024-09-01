package domain

import "time"

type User struct {
	Id       int64
	Email    string // 邮箱
	Password string // 密码

	Nickname string    // 用户名
	Birthday time.Time // 生日
	AboutMe  string    // 自己介绍

	Phone string // 手机
	Ctime time.Time
}
