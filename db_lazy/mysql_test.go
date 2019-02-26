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
		orm.Sync(&OperateLogTb{})
		lazyMysql = NewLazyMysql(orm, 30)
		go lazyMysql.Exec()
		time.Sleep(10*time.Millisecond)
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
	for i := 1; i < 10; i++ {
		lazyMysql.Add(&OperateLogTb{
			AdminName:      fmt.Sprintf("AdminName%d", i),
			OperateContent: fmt.Sprintf("OperateContent%d", i),
			OperateType:    0,
			CreatedTime:    time.Now().Unix(),
		}, LazyOperateType_Insert, nil, "")
	}
	time.Sleep(1 * time.Second)
	for i := 10; i < 20; i++ {
		lazyMysql.Add(&OperateLogTb{
			AdminName:      fmt.Sprintf("AdminName%d", i),
			OperateContent: fmt.Sprintf("OperateContent%d", i),
			OperateType:    1,
			CreatedTime:    time.Now().Unix(),
		}, LazyOperateType_Insert, nil, "")
	}
	time.Sleep(1 * time.Second)
	for i := 20; i < 24; i++ {
		err := lazyMysql.AddSQL("INSERT INTO operate_log_tb(admin_name,operate_content,operate_type,created_time) "+
			"value('%s','%s','%d','%d');", fmt.Sprintf("AdminName%d", i), fmt.Sprintf("OperateContent%d", i), 2, time.Now().Unix())
		if err != nil {
			t.Logf("err:%v", err)
		}
	}
	for i := 20; i < 24; i++ {
		err := lazyMysql.AddSQL("UPDATE operate_log_tb set operate_content='%s' where admin_name='%s';", fmt.Sprintf("OperateContent_%d", i), fmt.Sprintf("AdminName%d", i))
		if err != nil {
			t.Logf("err:%v", err)
		}
	}
	lazyMysql.Quit()
}
