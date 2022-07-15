package biz

import (
	"context"
	"user/internal/data/entity"

	"github.com/go-kratos/kratos/v2/log"
)

type User struct {
	Id string
	Name string
	Email string
	Role string
}

type UserFields struct {
	Id string
	Name string
	Email string
}

type UserRepo interface {
	UpdateUser(ctx context.Context, id string, user *entity.User) error
}

type UserUsecase struct {
	repo UserRepo
	log  *log.Helper
}

func NewUserUsecase(repo UserRepo, logger log.Logger) *UserUsecase {
	return &UserUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *UserUsecase) UpdateUser(ctx context.Context, newUser *User) error {
	uc.log.WithContext(ctx).Infof("update user id: %v", newUser.Id)

	err := uc.repo.UpdateUser(ctx, newUser.Id, NewUserPo(newUser).ToPo())

	if err != nil {
		uc.log.WithContext(ctx).Errorf("update user id: %v, err: %v", newUser.Id, err)
		return err
	}

	return nil
}
