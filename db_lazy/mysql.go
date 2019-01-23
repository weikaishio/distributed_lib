package db_lazy

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/mkideal/log"
	"github.com/weikaishio/distributed_lib"
)

type LazyOperateType int

const (
	LazyOperateType_Insert LazyOperateType = 1
	LazyOperateType_Update LazyOperateType = 2
	LazyOperateType_Delete LazyOperateType = 3
)

type LazyMysqlOperate struct {
	seq         int32
	tb          interface{}
	operateType LazyOperateType
	cols        []string
	condition   string // 1=1 and x=xx
	limit       int
}

type operatesSupportSort []*LazyMysqlOperate

func (o operatesSupportSort) Len() int {
	return len(o)
}

func (o operatesSupportSort) Less(i, j int) bool {
	return o[i].seq < o[j].seq
}

func (o operatesSupportSort) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

type LazyMysql struct {
	ormEngine    *xorm.Engine
	waitHandle   map[int32]*LazyMysqlOperate
	lock         sync.RWMutex
	seq          int32
	isRunning    int32
	quit         chan struct{}
	lazyDuration time.Duration
}

func NewLazyMysql(orm *xorm.Engine, lazyTimeSecond int) *LazyMysql {
	machineDB := &LazyMysql{
		ormEngine:    orm,
		waitHandle:   make(map[int32]*LazyMysqlOperate),
		lock:         sync.RWMutex{},
		quit:         make(chan struct{}),
		lazyDuration: time.Second * time.Duration(lazyTimeSecond),
	}
	return machineDB
}

var (
	ERR_AlreadyStop               = errors.New("lazy_mysql already stop")
	ERR_NotImpletementOperateType = errors.New("lazy_mysql not implatement type")
)

/*
Done:执行顺序需要确定 +排序
执行完成后删除元素
Done:锁粒度降低
todo:未做每次flush的执行数量限制, 暂不用
*/
func (a *LazyMysql) Quit() {
	log.Warn("LazyMysql's Quit begin")
	atomic.StoreInt32(&a.isRunning, 0)
	a.quit <- struct{}{}
	a.Flush() //再次确认处理完毕
	log.Warn("LazyMysql's Quit end")
}
func (a *LazyMysql) IsRunning() bool {
	return atomic.LoadInt32(&a.isRunning) == 1
}
func (a *LazyMysql) Exec() {
	log.Warn("LazyMysql's Exec begin")
	if !atomic.CompareAndSwapInt32(&a.isRunning, 0, 1) {
		log.Error("LazyMysql's Exec is already running, stop the Exec")
		return
	}
	defer atomic.CompareAndSwapInt32(&a.isRunning, 1, 0)
	for a.IsRunning() {
		distributed_lib.Try(func() {
			err := a.Flush()
			if err != nil {
				log.Error("LazyMysql's Exec Flush err:%v", err)
			}
		})
		select {
		case <-time.After(a.lazyDuration):
		case <-a.quit:
			log.Warn("LazyMysql's Exec end")
			break
		}
	}
}
func (a *LazyMysql) Flush() error {
	a.lock.RLock()
	if len(a.waitHandle) == 0 {
		a.lock.RUnlock()
		log.Trace("LazyMysql's Flush need handle count is 0")
		return nil
	}

	log.Trace("LazyMysql's Flush begin")
	defer log.Trace("LazyMysql's Flush end")

	log.Trace("LazyMysql's Flush need handle count:%d", len(a.waitHandle))
	keys := make([]int32, 0)

	var operates operatesSupportSort
	for _, v := range a.waitHandle {
		operates = append(operates, v)
	}
	a.lock.RUnlock()

	sort.Sort(operates)

	session := a.ormEngine.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return fmt.Errorf("lazy_mysql's Flush session.Begin err:%v", err)
	}
	for _, v := range operates {
		log.Trace("LazyMysql's Flush handle seq:%d", v.seq)
		switch v.operateType {
		case LazyOperateType_Insert:
			_, err = session.InsertOne(v.tb)
			if err != nil {
				log.Error("lazy_mysql's Flush Insert data:%v,condition:%v,err:%v", v.tb, v.condition, err)
				continue //session no need rollback
			}
		case LazyOperateType_Update:
			if v.condition == "" {
				v.condition = "1=1"
			}
			session = session.Where(v.condition).Cols(v.cols...)
			if v.limit > 0 {
				session.Limit(v.limit)
			}
			_, err = session.Update(v.tb)
			if err != nil {
				log.Error("lazy_mysql's Flush Update data:%v,condition:%v,v.cols:%v,err:%v", v.tb, v.condition, v.cols, err)
				continue
			}
		case LazyOperateType_Delete:
			if v.condition != "" {
				session = session.Where(v.condition).Cols(v.cols...)
				if v.limit > 0 {
					session.Limit(v.limit)
				}
				_, err = session.Delete(v.tb)
				if err != nil {
					log.Error("lazy_mysql's Flush Delete data:%v,condition:%v,v.cols:%v,err:%v", v.tb, v.condition, v.cols, err)
					continue
				}
			} else {
				log.Error("lazy_mysql's Flush Delete data:%v,condition:%v is nil not support", v.tb, v.condition, v.cols, err)
			}
		default:
			tbBys, _ := json.Marshal(v.tb)
			log.Error("LazyMysql's Flush %v:%d,data:%v,condition:%v", ERR_NotImpletementOperateType, v.operateType, string(tbBys), v.condition)
		}
		keys = append(keys, v.seq)
	}
	err = session.Commit()
	if err != nil {
		session.Rollback()
		return fmt.Errorf("lazy_mysql's Flush session.Commit err:%v", err)
	} else {
		a.lock.Lock()
		defer a.lock.Unlock()
		for _, k := range keys {
			delete(a.waitHandle, k)
		}
	}
	return nil
}

//if it added fail, need handle directly
func (a *LazyMysql) Add(tb interface{}, operateType LazyOperateType, cols []string, condition string) error {
	return a.AddWithLimit(tb, operateType, cols, condition, 0)
}
func (a *LazyMysql) AddWithLimit(tb interface{}, operateType LazyOperateType, cols []string, condition string, limit int) error {
	if atomic.LoadInt32(&a.isRunning) != 1 {
		log.Error("LazyMysql's Add exec failed, for it's already stoped (%v,%v,%v)", tb, operateType, condition)
		return ERR_AlreadyStop
	}
	atomic.AddInt32(&a.seq, 1)
	operateObj := &LazyMysqlOperate{
		seq:         atomic.LoadInt32(&a.seq),
		tb:          tb,
		operateType: operateType,
		cols:        cols,
		condition:   condition,
		limit:       limit,
	}
	a.lock.Lock()
	a.waitHandle[operateObj.seq] = operateObj
	a.lock.Unlock()

	log.Trace("LazyMysql's Add exec seq:%d", operateObj.seq)
	return nil
}
