package data

import (
	linq "github.com/ahmetb/go-linq/v3"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"user/internal/conf"
	"user/internal/data/entity"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewUserRepo)

// Data .
type Data struct {
	MysqlDb *gorm.DB
	PgsqlDb *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	// mysqldsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	mysqldsn := c.Database.Mysqldsn
	mysqldb, err := gorm.Open(mysql.Open(mysqldsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	mysqldb.AutoMigrate(&entity.UserFields{})

	// pgsqldsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	pgsqldsn := c.Database.Pgsqldsn
	pgsqldb, err := gorm.Open(postgres.Open(pgsqldsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	pgsqldb.AutoMigrate(&entity.UserRole{})

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		MysqlDb: mysqldb,
		PgsqlDb: pgsqldb,
	}, cleanup, nil
}

func (d *Data) NewXid() string {
	return uuid.NewString()
}

func (data *Data) StartXA(infos []*XA) error {
	err := linq.From(infos).
		SelectT(func(info *XA) error {
			return info.StartXA()
		}).
		WhereT(func(err error) bool {
			return err != nil
		}).
		First().(error)
	return err
}

func (data *Data) EndXA(infos []*XA) error {
	err := linq.From(infos).
		SelectT(func(info *XA) error {
			return info.EndXA()
		}).
		WhereT(func(err error) bool {
			return err != nil
		}).
		First().(error)
	return err
}

func (data *Data) Prepare(infos []*XA) error {
	err := linq.From(infos).
		SelectT(func(info *XA) error {
			return info.Prepare()
		}).
		WhereT(func(err error) bool {
			return err != nil
		}).
		First().(error)
	return err
}

func (data *Data) Commit(infos []*XA) error {
	err := linq.From(infos).
		SelectT(func(info *XA) error {
			return info.Commit()
		}).
		WhereT(func(err error) bool {
			return err != nil
		}).
		First().(error)
	return err
}

func (data *Data) Rollback(infos []*XA) error {
	err := linq.From(infos).
		SelectT(func(info *XA) error {
			return info.Rollback()
		}).
		WhereT(func(err error) bool {
			return err != nil
		}).
		First().(error)
	return err
}

func (target *Data) Run(xas []*XA) error {
	err := target.StartXA(xas)
	if err != nil {
		return err
	}
	err = target.EndXA(xas)
	if err != nil {
		return err
	}
	err = target.Prepare(xas)
	if err != nil {
		err := target.Rollback(xas)
		return err
	}
	err = target.Commit(xas)
	if err != nil {
		return err
	}
	return nil
}
