package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Password string
	Phone    string
	BirthDay string
	NickName string
	Profile  string
	Ctime    time.Time
}
