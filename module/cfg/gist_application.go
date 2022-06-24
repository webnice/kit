// Package cfg
package cfg

import (
	"errors"
	"fmt"
	"os"
	runtimeDebug "runtime/debug"
	"strings"

	kitTypes "github.com/webnice/kit/v3/types"
)

// Завершение приложения с кодом ошибки
func (essence *gist) exitWithCode() {
	var (
		ewc  kitTypes.ErrorWithCode
		tmp  string
		code uint8
		n    int
		ok   bool
	)

	for n = range essence.parent.error {
		if tmp = strings.TrimSpace(essence.parent.error[n].Error()); tmp != "" {
			_, _ = fmt.Fprintln(essence.parent.rawWriter, tmp)
		}
		if ewc, ok = essence.parent.error[n].(kitTypes.ErrorWithCode); ok {
			if ewc.Code() > code {
				code = ewc.Code()
			}
		}
	}
	if code > 0 {
		os.Exit(int(code))
	}
}

// Безопасный запуск главной функции приложения
func (essence *gist) safeLaunchApplication() {
	var (
		code uint8
		err  error
	)

	defer func() {
		if e := recover(); e != nil {
			err = essence.parent.Errors().ApplicationPanicException(0, e, runtimeDebug.Stack())
			essence.parent.error = append(essence.parent.error, err)
		}
	}()
	// Обеспечение выполнения одной копии приложения
	// При ошибке или попытке вызова приложения из приложения, случится паника с исключением "deadlock":
	// fatal error: all goroutines are asleep - deadlock!
	essence.parent.main.FnMutex.Lock()
	defer essence.parent.main.FnMutex.Unlock()
	// Выполнение основной функции с возвратом кода ошибки
	if code, err = essence.parent.main.Fn(); code != 0 || err != nil {
		if err == nil {
			err = errors.New("")
		}
		if code == 0 {
			essence.parent.error = append(essence.parent.error, essence.parent.Errors().ApplicationUnknownError(0, err))
		} else {
			essence.parent.error = append(essence.parent.error, kitTypes.NewErrorWithCode(code, err.Error()))
		}
	}

	return
}
