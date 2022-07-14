package service

import (
	"context"
	v1 "user/api/user/v1"
	"user/internal/biz"
	"user/internal/service/todo"

	"google.golang.org/protobuf/types/known/emptypb"
)

type UserService struct {
	v1.UnimplementedUserServiceServer

	uc *biz.UserUsecase
}

func NewUserService(uc *biz.UserUsecase) *UserService {
	return &UserService{uc: uc}
}

func (s *UserService) UpdateUser(ctx context.Context, in *v1.UpdateUserReq) (*emptypb.Empty, error) {
	err := s.uc.UpdateUser(ctx, todo.NewUserDto(in).ToDo())
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

