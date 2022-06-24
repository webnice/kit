// Package application
package application

import kitModuleCfg "github.com/webnice/kit/v3/module/cfg"

const (
	tplApplicationBegin    = "Приложение начало инициализацию."
	tplApplicationFinalize = "Приложение начало завершение компонентов."
	tplApplicationEnd      = "Приложение завершило работу."
	tplIsAccumulatedError  = "На уровне выполнения приложения %d, накопились ошибки."
)

// Единичный экземпляр объекта пакета.
var singleton *impl

// Interface is an interface of package.
type Interface interface {
	// Main Точка запуска, выполнения и завершения приложения.
	// Функция возвращает код ошибки, который передаётся в операционную систему и может быть считан запускающим
	// приложением, скриптом или операционной системой.
	Main() (code uint8, err error)

	// Cfg Возвращает интерфейс конфигурации приложения.
	Cfg() kitModuleCfg.Interface
}

// Объект сущности, интерфейс Interface.
type impl struct {
	finalize chan struct{}          // Канал ожидания окончания выполнения Finalize() перед завершением приложения.
	cfg      kitModuleCfg.Interface // Интерфейс конфигурации приложения.
}
