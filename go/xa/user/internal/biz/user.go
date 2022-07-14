package biz

import (
	"context"

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
	UpdateUserRole(ctx context.Context, id string, role string) error
	UpdateFields(ctx context.Context, id string, userFields *UserFields) error
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

	newRole := newUser.Role

	if newRole != "" {
		err := uc.repo.UpdateUserRole(ctx, newUser.Id, newRole)
		if err != nil {
			return err
		}
		err = uc.repo.UpdateFields(ctx, newUser.Id, &UserFields{
			Id: newUser.Id,
			Name: newUser.Name,
			Email: newUser.Email,
		})
		return err
	} else {
		err := uc.repo.UpdateFields(ctx, newUser.Id, &UserFields{
			Id: newUser.Id,
			Name: newUser.Name,
			Email: newUser.Email,
		})
		return err
	}
}
