package db_lazy

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/mkideal/log"
)

var (
	orm       *xorm.Engine
	lazyMysql *LazyMysql
)

func init() {
	driver := "mysql"
	host := "127.0.0.1:3306"
	database := "bg_db"
	username := "root"
	password := ""
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&&allowOldPasswords=1&parseTime=true", username, password, host, database)

	var err error
	orm, err = xorm.NewEngine(driver, dataSourceName)
	if err != nil {
		log.Fatal("NewEngine:%s,err:%v", dataSourceName, err)
	} else {
		lazyMysql = NewLazyMysql(orm, 30)
		go lazyMysql.Exec()
	}
}

type OperateLogTb struct {
	Id             int    `xorm:"not null pk autoincr INT(11)"`
	AdminName      string `xorm:"VARCHAR(30)"`
	OperateContent string `xorm:"VARCHAR(1000)"`
	OperateType    int    `xorm:"default 0 TINYINT(4)"`
	CreatedTime    int64  `xorm:"BIGINT(20)"`
}

func TestLazyMysql_Add(t *testing.T) {
	for i := 1000; i < 1010; i++ {
		lazyMysql.Add(&OperateLogTb{
			AdminName:      fmt.Sprintf("AdminName%d", i),
			OperateContent: fmt.Sprintf("OperateContent%d", i),
			OperateType:    0,
			CreatedTime:    time.Now().Unix(),
		}, LazyOperateType_Insert, nil, "")
	}
	time.Sleep(35 * time.Second)
	for i := 1100; i < 1110; i++ {
		lazyMysql.Add(&OperateLogTb{
			AdminName:      fmt.Sprintf("AdminName%d", i),
			OperateContent: fmt.Sprintf("OperateContent%d", i),
			OperateType:    1,
			CreatedTime:    time.Now().Unix(),
		}, LazyOperateType_Insert, nil, "")
	}
	lazyMysql.Quit()
}
