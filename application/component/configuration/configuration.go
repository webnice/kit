// Package configuration
package configuration

import (
	"bytes"
	"fmt"
	"os"

	kitModuleCfg "github.com/webnice/kit/v3/module/cfg"
	kitModuleCfgReg "github.com/webnice/kit/v3/module/cfg/reg"
	kitTypes "github.com/webnice/kit/v3/types"
)

// Регистрация компоненты в приложении.
func init() { kitModuleCfgReg.Registration(newComponent()) }

// Конструктор объекта компоненты.
func newComponent() kitTypes.Component { return &impl{cfg: kitModuleCfg.Get(), cmd: new(cmd)} }

// Ссылка на функцию получения значения режима отладки, для удобного использования внутри компоненты.
func (ccf *impl) debug() bool { return ccf.cfg.Debug() }

// Ссылка на менеджер логирования, для удобного использования внутри компоненты.
func (ccf *impl) log() kitTypes.Logger { return ccf.cfg.Log() }

// Preferences Функция возвращает настройки компоненты.
func (ccf *impl) Preferences() kitTypes.ComponentPreferences {
	const cEnvironment = `(?mi)application/component/environment$`
	return kitTypes.ComponentPreferences{
		After: []string{cEnvironment},
		Command: []kitTypes.ComponentCommand{
			{}, // Компонента будет запускаться как для своих команд, так и для любой другой команды.
			{
				Command:          "config",
				Description:      "Работа с конфигурационным файлом приложения.",
				Value:            ccf.cmd,
				GroupKey:         "main",
				GroupTitle:       "Основные режимы работы:",
				GroupDescription: "Команды основных режимов работы приложения.",
			},
		},
	}
}

// Initiate Функция инициализации компонента и подготовки компонента к запуску.
func (ccf *impl) Initiate() (err error) {
	var (
		appName    string
		fileName   string
		dirConf    string
		dirHome    string
		dirWork    string
		fi         os.FileInfo
		bufContent []byte
		buf        *bytes.Buffer
		cmd        []string
	)

	// Инициализация конфигурационного файла приложения не выполняется для команд config create и config test
	if ccf.cfg.Command() == cmdConfig && len(ccf.cfg.CommandFull()) > 1 {
		switch cmd = ccf.cfg.CommandFull(); cmd[1] {
		case cmdCreate, cmdTest:
			return
		}
	}
	// Если не указан файл конфигурации, сбор списка папок и поиск файла конфигурации.
	if appName, dirConf, dirHome, dirWork, fileName =
		ccf.cfg.AppName(),
		ccf.cfg.DirectoryConfig(),
		ccf.cfg.DirectoryHome(),
		ccf.cfg.DirectoryWorking(),
		ccf.cfg.FileConfig(); fileName == "" {
		// Поиск файла конфигурации по всем предопределённым директориям. Директории платформо-зависимые.
		fileName = ccf.findConfigurationFile(appName, dirConf, dirHome, dirWork)
	}
	// Если файл найден, установка значения пути и имени файла в конфигурации.
	if fileName != "" {
		ccf.cfg.Gist().FileConfig(fileName)
	}
	// Файл указан или найден, работа с файлом, если файла нет, выход.
	if fileName = ccf.cfg.FileConfig(); fileName == "" {
		return
	}
	// Проверка наличия файла
	if fi, err = os.Stat(fileName); err != nil {
		switch {
		case os.IsNotExist(err):
			err = ccf.cfg.Errors().ConfigurationFileNotFound(0, fileName, err)
		case os.IsPermission(err):
			err = ccf.cfg.Errors().ConfigurationPermissionDenied(0, fileName, err)
		default:
			err = ccf.cfg.Errors().ConfigurationUnexpectedMistakeFileAccess(0, fileName, err)
		}
		return
	}
	if fi.IsDir() {
		err = ccf.cfg.Errors().ConfigurationFileIsDirectory(0, fileName)
		return
	}
	// Чтение файла конфигурации в память.
	if bufContent, err = os.ReadFile(fileName); err != nil {
		err = ccf.cfg.Errors().ConfigurationFileReadingError(0, fileName, err)
		return
	}
	buf = bytes.NewBuffer(bufContent)
	// Принудительная очистка занятой памяти буферами, после завершения работы функции.
	defer func() { buf.Reset(); bufContent = nil }()
	// Создание общей структуры конфигурации, состоящей из стартовой структуры приложения и всех имплантированных
	// конфигураций компонентов. Выполнение загрузки данных и копирование во все имплантированные структуры.
	if err = ccf.cfg.Gist().ConfigurationLoad(buf); err != nil {
		switch err.(type) {
		case kitTypes.ErrorWithCode:
			// Ошибка уже стандартизирована.
		default:
			// Ошибка не стандартизированная
			err = ccf.cfg.Errors().ConfigurationApplicationObject(0, err)
		}
		return
	}
	if ccf.debug() {
		ccf.log().Infof(tplConfigurationInitEnd)
	}

	return
}

// Do Выполнение компонента приложения.
func (ccf *impl) Do() (levelDone bool, levelExit bool, err error) {
	var (
		wd  string
		cmd []string
	)

	// Переход в рабочую директорию приложения.
	if wd = ccf.cfg.DirectoryWorking(); wd != "" {
		if err = ccf.cfg.DirectoryWorkingChdir(); err != nil {
			err = ccf.cfg.Errors().CantChangeWorkDirectory(0, err)
			return
		}
	}
	if cmd = ccf.cfg.CommandFull(); ccf.cfg.Command() == cmdConfig && len(cmd) > 1 {
		switch cmd[1] {
		case cmdCreate:
			levelDone, levelExit, err = true, true, ccf.Create()
		case cmdTest:
			levelDone, levelExit, err = true, true, ccf.Test()
		default:
			err = fmt.Errorf(tplCommandNotImplemented, cmd[1])
		}
		if err != nil {
			return
		}
	}
	// В режиме отладки, печать конфигурационного файла в лог приложения.
	if ccf.debug() && !levelExit {
		ccf.log().Debug(ccf.cfg.ConfigurationUnionSprintf())
	}

	return
}

// Finalize Функция вызывается перед завершением компонента и приложения в целом.
func (ccf *impl) Finalize() (err error) { return }
