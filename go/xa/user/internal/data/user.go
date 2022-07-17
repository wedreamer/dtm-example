package data

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/dtm-labs/dtmcli"

	"user/internal/biz"
	"user/internal/data/entity"
	"user/internal/utils"

	"github.com/go-kratos/kratos/v2/log"
)

var T = utils.T

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

func (r *userRepo) UpdateUser(ctx context.Context, id string, user *entity.User) error {
	role := user.Role
	if role == "" {
		_, err := r.updateUserFields(ctx, id, &entity.UserFields{
			Id:    id,
			Name:  user.Name,
			Email: user.Email,
		}, false)
		return err
	} else {
		// use XA
		// generate sql
		pgsqlStr, err := r.updateUserRole(ctx, id, role, true)
		if err != nil {
			return err
		}
		mysqlStr, err := r.updateUserFields(ctx, id, &entity.UserFields{
			Id:    id,
			Name:  user.Name,
			Email: user.Email,
		}, true)
		if err != nil {
			return err
		}
		// use dtm XA
		gid := r.data.GenXid()
		pgXaInfo := &LocalTransactionInfo{
			DbType: PGSQL,
			CallBack: func(db *sql.DB, xa *dtmcli.Xa) error {
				_, err := db.Exec(pgsqlStr)
				return err
			},
		}
		mysqlXaInfo := &LocalTransactionInfo{
			DbType: MYSQL,
			CallBack: func(db *sql.DB, xa *dtmcli.Xa) error {
				_, err := db.Exec(mysqlStr)
				return err
			},
		}
		err = r.data.StartXa(gid, []*LocalTransactionInfo{pgXaInfo, mysqlXaInfo})
		if err != nil {
			return err
		}
		/* xid := r.data.NewXid()
		pgXa := &XA{}
		pgXa.SetXId(xid)
		pgXa.SetSql(pgsqlStr, PGSQL)
		pgXa.Init(r.data.PgsqlDb, PGSQL)
		mysqlXa := &XA{}
		mysqlXa.SetXId(xid)
		mysqlXa.SetSql(mysqlStr, MYSQL)
		mysqlXa.Init(r.data.MysqlDb, MYSQL)
		xas := []*XA{pgXa, mysqlXa}
		err = r.data.Run(xas)
		if err != nil {
			return err
		} */
	}
	return nil
}

func (r *userRepo) updateUserRole(ctx context.Context, id string, role string, onlySql bool) (string, error) {
	if onlySql {
		// TODO: sql build -> dbr squirrel sqrl
		sql := sq.Update((&entity.UserRole{}).TableName()).Where("id = ?", id).Set("role = ?", role)
		sqlstr, _, err := sql.ToSql()
		return sqlstr, err
	}
	res := r.data.PgsqlDb.Model(&entity.UserRole{}).Where("id = ?", id).Updates(map[string]interface{}{"role": role})
	return "", res.Error
}

func (r *userRepo) updateUserFields(ctx context.Context, id string, userFields *entity.UserFields, onlySql bool) (string, error) {
	if onlySql {
		sql := sq.Update((&entity.UserFields{}).TableName()).Where("id = ?", id).Set("name = ?", userFields.Name).Set("email = ?", userFields.Email)
		sqlstr, _, err := sql.ToSql()
		return sqlstr, err
	}
	res := r.data.MysqlDb.Model(&entity.UserFields{}).Where("id = ?", id).Updates(userFields)
	return "", res.Error
}
