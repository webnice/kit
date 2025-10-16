package loggerconsole

import (
	"sync"

	"github.com/webnice/kit/v4/plugin/component/logger_console/tpl"

	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
)

const (
	tplOnInitiate = "Запущена компонента вывода форматированных сообщений логирования в STDERR."
	tplOnFinalize = "Остановлена компонента вывода форматированных сообщений логирования в STDERR."
)

const defaultTpl = `${dye:all:level}${-spc-}

${#: Дата и время с форматом вывода: }${#:timestamp:Europe/Moscow:02.01.2006 15:04:05.000000}

${-spc-}

[${level:S:1}:${level:d:1}]:${dye:reset:all} ${message}${bp--}

${-spc-:1}           ${dye:text:black:bright}{${package}/${dye:set:Underline}${shortfile}${dye:reset:Underline}:${dye:set:reverse}${line}${dye:reset:reverse},

${-spc-:1}  функция: 

${-spc-:1}           ${dye:back:Black:normal}${dye:set:Underline}${function}()${dye:reset:all}${dye:text:black:bright}}${dye:reset:all}

${-spc-:1}${-spc-}

${#: Ключи: }${-spc-:1}${dye:text:#0087FF}${keys}`

// Interface Интерфейс пакета.
type Interface interface {
}

// Объект сущности пакета.
type impl struct {
	err            error
	cfg            kitModuleCfg.Interface
	tpl            tpl.Interface
	handlerMux     *sync.Mutex // Блокировка обработки сообщений до выполнения функции инициализации.
	handlerControl bool        // Включение контроля через блокировку.
}
