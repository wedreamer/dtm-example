package data

import (
	"context"

	"user/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
)

type userRepo struct {
	data *Data
	log  *log.Helper
}

func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *userRepo) UpdateUserRole(ctx context.Context, id string, role string) error {
	return nil
}

func (r *userRepo) UpdateFields(ctx context.Context, id string, userFields *biz.UserFields) error {
	return nil
}

