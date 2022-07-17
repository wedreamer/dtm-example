package data

import (
	"database/sql"
	"user/internal/conf"
	"user/internal/data/entity"

	linq "github.com/ahmetb/go-linq/v3"
	"github.com/dtm-labs/dtmcli"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewUserRepo)

// Data .
type Data struct {
	MysqlDb *gorm.DB
	PgsqlDb *gorm.DB
	GenXid 	func() string
	DtmServer string
	StartXa func(gid string, localTraInfos []*LocalTransactionInfo) error
}

type LocalTransactionInfo struct {
	DbType DbType
	CallBack func(db *sql.DB, xa *dtmcli.Xa) error
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

	dtmServer := T("%s:%d", c.Database.DtmServer.Host, c.Database.DtmServer.Port)

	startXa := func(gid string, localTraInfos [] *LocalTransactionInfo) error {
		err := dtmcli.XaGlobalTransaction(dtmServer, gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
			// 本地事务执行
			err := linq.From(localTraInfos).SelectT(func(info *LocalTransactionInfo) error {
				dbconf := dtmcli.DBConf{}
				if info.DbType == MYSQL {
					dbconf.Driver = "mysql"
					dbconf.Host = "mysql"
					dbconf.Port = 3306
					dbconf.User = "pass"
					dbconf.Password = "localhost"
					dbconf.Db = "test"
				} else if info.DbType == PGSQL {
					dbconf.Driver = "postgres"
					dbconf.Host = "localhost"
					dbconf.Port = 9920
					dbconf.User = "gorm"
					dbconf.Password = "gorm"
					dbconf.Db = "gorm"
				}
				err := dtmcli.XaLocalTransaction(nil, dbconf, info.CallBack)
				if err != nil {
					return err
				}
				return err
			}).
			First().
			(error)
			// 有一个有异常, 则返回错误
			return nil, err
		})
		return err
	}

	return &Data{
		MysqlDb: mysqldb,
		PgsqlDb: pgsqldb,
		DtmServer: dtmServer,
		GenXid: func() string { return dtmcli.MustGenGid(dtmServer) },
		StartXa: startXa,
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
