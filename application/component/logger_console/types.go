// Package loggerconsole
package loggerconsole

import (
	"sync"

	"github.com/webnice/kit/application/component/logger_console/tpl"

	kitModuleCfg "github.com/webnice/kit/module/cfg"
)

const (
	tplOnInitiate = "Запущена компонента вывода форматированных сообщений логирования в STDERR."
	tplOnFinalize = "Остановлена компонента вывода форматированных сообщений логирования в STDERR."
)

// :Europe/Moscow:02.01.2006 15:04:05.000000

const defaultTpl = `${dye:all:level}${-spc-}

${#: Дата и время с форматом вывода: }${#:timestamp:Europe/Moscow:02.01.2006 15:04:05.000000}
${-spc-}

[${level:S:1}:${level:d:1}]:${dye:all:reset} ${message}${bp--}

${-spc-:1}           ${dye:text:black:bright}{${package}/${shortfile}:${line}, функция: 
${-spc-:1}           ${dye:back:Black:normal}${function}()${dye:all:reset}${dye:text:black:bright}}${dye:all:reset}
${-spc-:1}
${-spc-}
${#: Ключи для примера: }${-spc-:1}${dye:text:#FF0955}${keys}`

// Interface is an interface of package.
type Interface interface {
}

// Структура объекта компоненты.
type impl struct {
	err            error
	cfg            kitModuleCfg.Interface
	tpl            tpl.Interface
	handlerMux     *sync.Mutex // Блокировка обработки сообщений до выполнения функции инициализации.
	handlerControl bool        // Включение контроля через блокировку.
}