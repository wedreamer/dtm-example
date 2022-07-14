package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"user/internal/conf"
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

	// pgsqldsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	pgsqldsn := c.Database.Pgsqldsn
	pgsqldb, err := gorm.Open(postgres.Open(pgsqldsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		MysqlDb: mysqldb,
		PgsqlDb: pgsqldb,
	}, cleanup, nil
}
