package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dtm-labs/dtmcli"
	"github.com/gin-gonic/gin"
	_ "github.com/go-redis/redis/v8"
	"github.com/go-resty/resty/v2"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

// 事务参与者的服务地址
const qsBusiAPI = "/api/busi_start"
const qsBusiPort = 8082

var qsBusi = fmt.Sprintf("http://localhost:%d%s", qsBusiPort, qsBusiAPI)

func main() {
	QsStartSvr()
	select {}
}

func QsStartSvr() {
	app := gin.New()
	qsAddRoute(app)
	log.Printf("quick start examples listening at %d", qsBusiPort)
	go func() {
		_ = app.Run(fmt.Sprintf(":%d", qsBusiPort))
	}()
	time.Sleep(100 * time.Millisecond)
}

func qsAddRoute(app *gin.Engine) {

	app.GET("/", func(context *gin.Context) {
		num := context.Query("num")
		numInt, _ := strconv.Atoi(num)
		gid, err := QsFireRequest(numInt)
		if err != nil {
			context.JSON(500, gin.H{"gid": gid, "num": num, "msg": err.Error()})
			return
		}
		context.JSON(200, gin.H{"gid": gid, "num": num, "msg": "ok"})
	})

	// 微服务1 入账接口
	app.POST(qsBusiAPI+"/TransIn", func(c *gin.Context) {
		err := dtmcli.XaLocalTransaction(c.Request.URL.Query(), dtmcli.DBConf{Driver: "mysql", Host: "0.0.0.0", Port: 3306, Db: "dtm_barrier", User: "root", Password: ""}, func(db *sql.DB, xa *dtmcli.Xa) error {
			bodyBytes, err := ioutil.ReadAll(c.Request.Body)
			reqBody := new(Req)
			_ = json.Unmarshal(bodyBytes, reqBody)
			_, err = db.Exec("update book set num = num + ? where id = 1", reqBody.Num)
			if err != nil {
				return err
			}
			log.Printf("TransIn")
			return nil
		})
		if err != nil {
			c.JSON(409, err.Error())
		} else {
			c.JSON(200, "")
		}
	})
	// 微服务二 出账接口
	app.POST(qsBusiAPI+"/TransOut", func(c *gin.Context) {
		err := dtmcli.XaLocalTransaction(c.Request.URL.Query(), dtmcli.DBConf{Driver: "mysql", Host: "0.0.0.0", Port: 3306, Db: "blog", User: "root", Password: ""}, func(db *sql.DB, xa *dtmcli.Xa) error {
			bodyBytes, err := ioutil.ReadAll(c.Request.Body)
			reqBody := new(Req)
			_ = json.Unmarshal(bodyBytes, reqBody)
			if reqBody.Num > 1000 {
				return errors.New("金额过大")
			}
			_, err = db.Exec("update book set num = num - ? where id = 2", reqBody.Num)
			if err != nil {
				return err
			}
			log.Printf("TransOut")
			return nil
		})
		if err != nil {
			c.JSON(409, err.Error())
		} else {
			c.JSON(200, "")
		}

	})
}

const dtmServer = "http://localhost:36789/api/dtmsvr"

type Req struct {
	Num int `json:"num"`
}

func QsFireRequest(amount int) (string, error) {
	req := &gin.H{"num": amount} // 微服务的载荷
	gid := dtmcli.MustGenGid(dtmServer)
	err := dtmcli.XaGlobalTransaction(dtmServer, gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		branch, err := xa.CallBranch(req, qsBusi+"/TransIn")
		if err != nil {
			return branch, err
		}
		return xa.CallBranch(req, qsBusi+"/TransOut")
	})
	log.Printf("transaction: %s submitted", gid)
	if err != nil {
		return gid, err
	}
	return gid, nil
}
