package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole struct {
	Id string `gorm:"primaryKey"`
	Role string `gorm:"not null;type:char(20)"`
}

func (user *UserRole) BeforeCreate(scope *gorm.DB) (err error) {
 // UUID version 4
  user.Id = uuid.NewString()
  return
}

func (user *UserRole) TableName() string {
	return "user_role"
}
