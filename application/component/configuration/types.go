// Package configuration
package configuration

import kitModuleCfg "github.com/webnice/kit/module/cfg"

const (
	osWindows       = "windows"
	symbolSlash     = `/`
	symbolBackslash = `\`
	pathParent      = `..`
	extensionYaml   = `.yml`
)

const (
	cmdConfig = "config"
	cmdCreate = "create"
	cmdTest   = "test"
)

const (
	tplConfigurationInitEnd  = "Завершена инициализация конфигурации."
	tplCommandNotImplemented = "команда %q не реализована"
)

// Структура объекта компоненты.
type impl struct {
	cfg kitModuleCfg.Interface
	cmd *cmd
}

type cmd struct {
	Create struct {
		Filename string `arg:"" help:"Название создаваемого файла."`
	} `cmd:"create" help:"Создание конфигурационного файла."`
	Test struct {
		Filename string `arg:"" help:"Название конфигурационного файла."`
	} `cmd:"test" help:"Тестирование конфигурационного файла."`
}
