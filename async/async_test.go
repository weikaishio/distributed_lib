package async

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"testing"
	"time"
)

var (
	engine *Engine
)

func init() {
	cpuCount := runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("cpu count:%d\n", cpuCount)
	engine = NewEngine(uint32(cpuCount))
}

func TestAsyncEngine_AddTask(t *testing.T) {
	go engine.Start()

	startTime := time.Now()
	for i := 0; i < 3; i++ {
		get("http://www.baidu.com")
		t.Logf("name: baidu, done\n")
		get("http://www.sina.com")
		t.Logf("name: sina, done\n")
		get("http://www.sohu.com")
		t.Logf("name: sohu, done\n")
	}
	t.Logf("single-goroutine use time: %v\n", time.Now().Sub(startTime).String())

	startTime = time.Now()

	var tasksWait []*Task
	for i := 0; i < 3; i++ {
		task, err := engine.AddTask(get, "http://www.baidu.com")
		if err != nil {
			t.Logf("addTask err:%v\n", err)
		} else {
			tasksWait = append(tasksWait, task)
		}
		task, err = engine.AddTask(get, "http://www.sina.com")
		if err != nil {
			t.Logf("addTask err:%v\n", err)
		} else {
			tasksWait = append(tasksWait, task)
		}
		task, err = engine.AddTask(get, "http://www.sohu.com")
		if err != nil {
			t.Logf("addTask err:%v\n", err)
		} else {
			tasksWait = append(tasksWait, task)
		}
	}
	time.Sleep(100 * time.Millisecond)

	go engine.Stop()

	for _, task := range tasksWait {
		select {
		case returnVal := <-task.ReturnVal:
			for i, val := range returnVal {
				switch typ := val.(type) {
				case string:
					t.Logf("resVal %d type string len:%v done \n", i, len(typ))
				case error:
					t.Logf("resVal %d type error:%v done \n", i, typ.Error())
				default:
					t.Logf("resVal %d type:%v done \n", i, reflect.TypeOf(val))
				}
			}
		}
	}

	t.Logf("async use time: %v\n", time.Now().Sub(startTime).String())

	_, err := engine.AddTask(get, "http://www.163.com")
	if err != nil {
		t.Logf("addTask err:%v\n", err)
	}
	t.Logf("TestAsyncEngine_AddTask end\n")
}

func get(url string) (string, error) {
	if rs, err := http.Get(url); err == nil {
		defer rs.Body.Close()
		if res, err := ioutil.ReadAll(rs.Body); err == nil {
			time.Sleep(100 * time.Millisecond)
			return string(res), errors.New("test error")
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}
