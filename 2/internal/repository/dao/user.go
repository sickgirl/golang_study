package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

// ErrDataNotFound 通用的数据没找到
var ErrDataNotFound = gorm.ErrRecordNotFound

// ErrUserDuplicate 这个算是 user 专属的
var ErrUserDuplicate = errors.New("用户邮箱或者手机号冲突")

type UserDAO interface {
	Insert(ctx context.Context, u User) error
	Update(ctx context.Context, nickName, profile string, birthDay string, id int64) error
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewGORMUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{
		db: db,
	}
}

func (ud *GORMUserDAO) Update(ctx context.Context, nickName, profile string, birthDay string, id int64) error {

	// 更新用户信息
	err := ud.db.WithContext(ctx).Where("id = ?", id).Updates(User{NickName: nickName, Profile: profile, BirthDay: birthDay}).Error
	if err != nil {
		return err
	}

	return nil
}

func (ud *GORMUserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := ud.db.WithContext(ctx).Create(&u).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const uniqueIndexErrNo uint16 = 1062
		if me.Number == uniqueIndexErrNo {
			return ErrUserDuplicate
		}
	}
	return err
}

func (ud *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).First(&u, "phone = ?", phone).Error
	return u, err
}

func (ud *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).First(&u, "email = ?", email).Error
	return u, err
}

func (ud *GORMUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).First(&u, "id = ?", id).Error
	return u, err
}

type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 设置为唯一索引
	Email    sql.NullString `gorm:"unique"`
	Password string

	//Phone *string
	Phone sql.NullString `gorm:"unique"`
	//生日
	BirthDay string `gorm:"column:birth_day;type:date" json:"birth_day"`
	//昵称
	NickName string `gorm:"column:nick_name;type:varchar(255)" json:"nick_name"`
	//简介
	Profile string `gorm:"column:profile;type:varchar(255)" json:"profile"`

	// 创建时间
	Ctime int64
	// 更新时间
	Utime int64
}
