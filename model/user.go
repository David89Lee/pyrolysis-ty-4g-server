package model

import (
	"errors"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/David89Lee/pyrolysis_ty_server/common"
)

type User struct {
	gorm.Model
	Username string `Gorm:"type:varchar(32); uniqueIndex; comment:'用户名'"`
	Password string `Gorm:"type:varchar(128); comment:'密码(bcrypt)'"`
	Nickname string `Gorm:"type:varchar(64); comment:'昵称'"`
}

// SeedAdmin 首次启动时创建默认管理员账号 admin
func SeedAdmin() {
	var count int64
	if err := DB.Model(&User{}).Count(&count).Error; err != nil {
		zap.L().Error("查询用户表失败", zap.Error(err))
		return
	}
	if count > 0 {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(common.AuthSetting.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		zap.L().Error("生成管理员密码失败", zap.Error(err))
		return
	}
	admin := &User{Username: "admin", Password: string(hash), Nickname: "系统管理员"}
	if err := DB.Create(admin).Error; err != nil {
		zap.L().Error("创建默认管理员失败", zap.Error(err))
		return
	}
	zap.L().Info("已创建默认管理员账号 admin")
}

// FindUserByUsername 按用户名查询用户
func FindUserByUsername(username string) (*User, error) {
	var user User
	err := DB.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CheckPassword 校验明文密码是否匹配
func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}
