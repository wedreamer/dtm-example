package todo

import (
	v1 "user/api/user/v1"
	biz "user/internal/biz"
)

type UserDto v1.UpdateUserReq

func NewUserDto(in *v1.UpdateUserReq) *UserDto {
	res := (*UserDto)(in)
	return res
}

func (target *UserDto) ToDo() *biz.User {
	return &biz.User{
		Id:   target.Id,
		Name: target.Name,
		Email: target.Email,
		Role: target.Role,
	}
}
