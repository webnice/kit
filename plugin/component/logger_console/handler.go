package loggerconsole

import (
	"bytes"
	"context"
	"fmt"

	"github.com/webnice/kit/v5/plugin/component/logger_console/tpl"

	kitModuleLog "github.com/webnice/kit/v5/module/log"
)

// Handler Обработчик сообщений лога.
func (lgc *impl) Handler(msg *kitModuleLog.Message) {
	var (
		err error
		ses tpl.Session
		buf *bytes.Buffer
	)

	// Ожидание завершения инициализации компоненты, либо остановки шины данных.
	select {
	case <-lgc.ready:
	case <-lgc.stopping():
		return
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

// Ссылка на канал уведомления об остановке потоков обработчика шины данных.
func (lgc *impl) stopping() (ret <-chan struct{}) {
	var ctx context.Context

	if ctx = lgc.cfg.Bus().Gist().Context(); ctx != nil {
		ret = ctx.Done()
	}

	return
}
