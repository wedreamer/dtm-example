package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserFields struct {
	Id string `gorm:"primaryKey"`
	Name string `gorm:"not null;type:char(20)"`
	Email string `gorm:"not null;type:char(50)"`
}

func (user *UserFields) BeforeCreate(scope *gorm.DB) (err error) {
 // UUID version 4
  user.Id = uuid.NewString()
  return
}

func (user *UserFields) TableName() string {
	return "user_fields"
}
