package configuration

import (
	"bytes"
	"fmt"
	"os"

	"github.com/webnice/dic"
	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitModuleCfgReg "github.com/webnice/kit/v4/module/cfg/reg"
	kitTypes "github.com/webnice/kit/v4/types"
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
	const cEnvironment = "(?mi)app/component/environment$"
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
		ierr       dic.IError
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
		// Поиск файла конфигурации по всем предопределённым директориям. Директории платформ-зависимые.
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
			err = ccf.cfg.Errors().ConfigurationFileNotFound.Bind(fileName, err)
		case os.IsPermission(err):
			err = ccf.cfg.Errors().ConfigurationPermissionDenied.Bind(fileName, err)
		default:
			err = ccf.cfg.Errors().ConfigurationUnexpectedMistakeFileAccess.Bind(fileName, err)
		}
		return
	}
	if fi.IsDir() {
		err = ccf.cfg.Errors().ConfigurationFileIsDirectory.Bind(fileName)
		return
	}
	// Чтение файла конфигурации в память.
	if bufContent, err = os.ReadFile(fileName); err != nil {
		err = ccf.cfg.Errors().ConfigurationFileReadingError.Bind(fileName, err)
		return
	}
	buf = bytes.NewBuffer(bufContent)
	// Принудительная очистка занятой памяти буферами, после завершения работы функции.
	defer func() { buf.Reset(); bufContent = nil }()
	// Создание общей структуры конфигурации, состоящей из стартовой структуры приложения и всех имплантированных
	// конфигураций компонентов. Выполнение загрузки данных и копирование во все имплантированные структуры.
	if err = ccf.cfg.Gist().ConfigurationLoad(buf); err != nil {
		if ierr = ccf.cfg.Errors().Unbind(err); ierr == nil {
			err = ccf.cfg.Errors().ConfigurationApplicationObject.Bind(err)
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
		tmp string
		cmd []string
	)

	// Переход в рабочую директорию приложения.
	if tmp = ccf.cfg.DirectoryWorking(); tmp != "" {
		if err = ccf.cfg.DirectoryWorkingChdir(); err != nil {
			err = ccf.cfg.Errors().CantChangeWorkDirectory.Bind(err)
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
		if tmp = ccf.cfg.ConfigurationUnionSprintf(); tmp != "" {
			ccf.log().Debug(tmp)
		}
	}

	return
}

// Finalize Функция вызывается перед завершением компонента и приложения в целом.
func (ccf *impl) Finalize() (err error) { return }
