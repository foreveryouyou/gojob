package utils

import (
	"errors"
	"fmt"
	"runtime"
)

// PanicToError Panic转换为error
func PanicToError(f func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(PanicTrace(e))
		}
	}()
	f()
	return
}

// PanicTrace panic调用链跟踪
func PanicTrace(err interface{}) string {
	stackBuf := make([]byte, 4096)
	n := runtime.Stack(stackBuf, false)

	return fmt.Sprintf("panic: %v %s", err, stackBuf[:n])
}
