package async

import (
	"errors"
	"reflect"
	"sync/atomic"
)

var (
	Err_AlreadyStop   = errors.New("The Engine has stopped")
	Err_InvalidParams = errors.New("The Handler param invalid")
)

type Task struct {
	ReturnVal chan []interface{}
	Handler   reflect.Value
	Params    []reflect.Value
}

type Engine struct {
	TaskQueue    chan *Task
	running      uint32
	quit         []chan struct{}
	routineCount int
}

func NewEngine(routineCount uint32) *Engine {
	return &Engine{
		TaskQueue:    make(chan *Task, 1000),
		routineCount: int(routineCount),
		quit:         make([]chan struct{}, 0),
	}
}
func (a *Engine) Start() {
	if !atomic.CompareAndSwapUint32(&a.running, 0, 1) {
		return
	}
	defer atomic.CompareAndSwapUint32(&a.running, 1, 0)

	for i := 0; i < a.routineCount; i++ {
		go func() {
			quit := make(chan struct{})
			a.quit = append(a.quit, quit)
			for {
				select {
				case <-quit:
					return
				case task := <-a.TaskQueue:
					values := task.Handler.Call(task.Params)

					if valuesNum := len(values); valuesNum > 0 {
						resultItems := make([]interface{}, valuesNum)
						for k, v := range values {
							resultItems[k] = v.Interface()
						}

						task.ReturnVal <- resultItems
					}
				}
			}
		}()
	}
	quit := make(chan struct{})
	a.quit = append(a.quit, quit)
	<-quit
}
func (a *Engine) AddTask(handler interface{}, params ...interface{}) (*Task, error) {
	if !a.IsRunning() {
		return nil, Err_AlreadyStop
	}
	handlerValue := reflect.ValueOf(handler)
	if handlerValue.Kind() != reflect.Func {
		return nil, Err_InvalidParams
	}

	paramNum := len(params)
	task := &Task{
		Handler:   handlerValue,
		Params:    make([]reflect.Value, paramNum),
		ReturnVal: make(chan []interface{}),
	}

	if paramNum > 0 {
		for k, v := range params {
			task.Params[k] = reflect.ValueOf(v)
		}
	}
	a.TaskQueue <- task
	return task, nil
}
func (a *Engine) IsRunning() bool {
	return atomic.LoadUint32(&a.running) != 0
}

//This method needs to be executed before the process exits
func (a *Engine) Stop() {
	if !a.IsRunning() {
		return
	}
	atomic.StoreUint32(&a.running, 0)
	for _, quit := range a.quit {
		quit <- struct{}{}
	}
}
