package loggerconsole

import (
	"bytes"
	"fmt"

	"github.com/webnice/kit/v4/app/component/logger_console/tpl"

	kitModuleLog "github.com/webnice/kit/v4/module/log"
)

// Handler Обработчик сообщений лога.
func (lgc *impl) Handler(msg *kitModuleLog.Message) {
	var (
		err error
		ses tpl.Session
		buf *bytes.Buffer
	)

	if lgc.handlerControl {
		lgc.handlerMux.Lock()
		defer lgc.handlerMux.Unlock()
	}
	if ses, err = lgc.tpl.
		NewSession(msg); err != nil {
		_, _ = fmt.Fprintln(lgc.wr(), err.Error())
		return
	}
	if buf, err = ses.Do(); err != nil {
		_, _ = fmt.Fprintln(lgc.wr(), err.Error())
		return
	}
	lgc.tpl.Output(buf)
}
