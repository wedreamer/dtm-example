package biz

import (
	"user/internal/data/entity"
)

type UserPo entity.User

func NewUserPo(in *User) *UserPo {
	res := (*UserPo)(in)
	return res
}

func (target *UserPo) ToPo() *entity.User {
	return &entity.User{
		Id:   target.Id,
		Name: target.Name,
		Email: target.Email,
		Role: target.Role,
	}
}
