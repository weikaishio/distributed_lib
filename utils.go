package distributed_lib

import (
	"fmt"
	"runtime"
	"errors"

	"github.com/mkideal/log"
)

func Try(fn func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 1<<16)
			buf = buf[:runtime.Stack(buf, true)]
			switch typ := e.(type) {
			case error:
				err = typ
			case string:
				err = errors.New(typ)
			default:
				err = fmt.Errorf("%v", typ)
			}
			log.Error("==== STACK TRACE BEGIN ====\npanic: %v\n%s\n===== STACK TRACE END =====", err, string(buf))
		}
	}()
	fn()
	return
}