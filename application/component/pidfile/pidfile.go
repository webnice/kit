// Package pidfile
package pidfile

import (
	"fmt"
	"os"
	"strings"

	kitModuleCfg "github.com/webnice/kit/v3/module/cfg"
	kitModuleCfgReg "github.com/webnice/kit/v3/module/cfg/reg"
	kitModuleLock "github.com/webnice/kit/v3/module/lock"
	kitTypes "github.com/webnice/kit/v3/types"
)

const (
	defaultModeFile       = 0644
	defaultModeDir        = 0755
	defaultFlag           = os.O_CREATE | os.O_RDWR | os.O_TRUNC | os.O_SYNC
	tplPidFileIsDirectory = "вместо имени PID файла указана директория: %s"
	tplPidContentRead     = "чтение содержимого файла %q, ошибка: %s"
	tplPidMake            = "создание PID файла %q, ошибка: %s"
	tplPidWrite           = "запись в PID файл %q идентификатора процесса, ошибка: %s"
	tplPidLock            = "блокировка PID файла %q, ошибка: %s"
	tplPidCreated         = "Успешно создан PID файл приложения %q."
	tplPidUnlock          = "снятие блокировки PID файла %q, ошибка: %s"
	tplPidClose           = "закрытие PID файла %q, ошибка: %s"
	tplPidDelete          = "удаление PID файла %q, ошибка: %s"
	tplPidEndFileDeleted  = "Завершена работы с PID файлом приложения, файл %q удалён."
	tplPidEnd             = "Завершена работы с PID файлом приложения."
)

// Структура объекта компоненты.
type impl struct {
	id    int
	pfnm  string
	pffh  *os.File
	cfg   kitModuleCfg.Interface
	lock  *kitModuleLock.Lock
	isRun bool
}

// Регистрация компоненты в приложении.
func init() { kitModuleCfgReg.Registration(newComponent()) }

// Конструктор объекта компоненты.
func newComponent() kitTypes.Component { return &impl{cfg: kitModuleCfg.Get()} }

// Ссылка на функцию получения значения режима отладки, для удобного использования внутри компоненты.
func (pid *impl) debug() bool { return pid.cfg.Debug() }

// Ссылка на менеджер логирования, для удобного использования внутри компоненты.
func (pid *impl) log() kitTypes.Logger { return pid.cfg.Log() }

// Preferences Функция возвращает настройки компоненты.
func (pid *impl) Preferences() kitTypes.ComponentPreferences {
	const (
		cConfiguration = `(?mi)application/component/configuration$`
		cLogging       = `(?mi)application/component/logging$`
	)
	return kitTypes.ComponentPreferences{
		After: []string{cConfiguration, cLogging},
	}
}

// Initiate Функция инициализации компонента и подготовки компонента к запуску.
func (pid *impl) Initiate() (err error) {
	pid.id = os.Getpid()
	return
}

// Do Выполнение компонента приложения.
func (pid *impl) Do() (levelDone bool, levelExit bool, err error) {
	if pid.isSkip() {
		return
	}
	if pid.pfnm = pid.cfg.FilePid(); pid.pfnm == "" {
		return
	}
	pid.isRun = true
	pid.createDirectoryForFile(pid.pfnm)
	if err = pid.IsProcessExist(); err != nil {
		levelDone, levelExit, pid.pfnm = true, true, ""
		return
	}
	// Создание, обрезка, открытие PID файла на чтение и запись.
	pid.pffh, err = os.OpenFile(pid.pfnm, defaultFlag, os.FileMode(defaultModeFile))
	if err != nil {
		levelDone, levelExit = true, true
		err = fmt.Errorf(tplPidMake, pid.pfnm, err)
		err = pid.cfg.Errors().PidFileError(0, pid.pfnm, err)
		return
	}
	// Запись в файл ID текущего процесса.
	if _, err = fmt.Fprintln(pid.pffh, pid.id); err != nil {
		levelDone, levelExit = true, true
		err = fmt.Errorf(tplPidWrite, pid.pfnm, err)
		err = pid.cfg.Errors().PidFileError(0, pid.pfnm, err)
		return
	}
	// Блокировка PID файла, как минимум попытка (блокировка в некоторых OS не работает).
	pid.lock = kitModuleLock.New(pid.pfnm)
	if err = pid.lock.Lock(); err != nil {
		levelDone, levelExit = true, true
		err = fmt.Errorf(tplPidLock, pid.pfnm, err)
		err = pid.cfg.Errors().PidFileError(0, pid.pfnm, err)
		return
	}
	if pid.cfg.Debug() && pid.pfnm != "" && pid.pffh != nil {
		pid.log().Infof(tplPidCreated, pid.pfnm)
	}

	return
}

// Finalize Функция вызывается перед завершением компонента и приложения в целом.
func (pid *impl) Finalize() (err error) {
	var isRm bool

	if pid.isSkip() {
		return
	}
	// Снятие блокировки с файла
	if pid.lock != nil && pid.lock.IsLocked() {
		if err = pid.lock.Unlock(); err != nil {
			err = fmt.Errorf(tplPidUnlock, pid.pfnm, err)
			pid.cfg.Gist().ErrorAppend(pid.cfg.Errors().PidFileError(0, pid.pfnm, err))
		}
	}
	// Закрытие файлового дескриптора
	if pid.pffh != nil {
		if err = pid.pffh.Close(); err != nil {
			err = fmt.Errorf(tplPidClose, pid.pfnm, err)
			pid.cfg.Gist().ErrorAppend(pid.cfg.Errors().PidFileError(0, pid.pfnm, err))
		}
	}
	// Удаление PID файла
	if pid.pfnm != "" {
		if err = os.Remove(pid.pfnm); err != nil {
			err = fmt.Errorf(tplPidDelete, pid.pfnm, err)
			pid.cfg.Gist().ErrorAppend(pid.cfg.Errors().PidFileError(0, pid.pfnm, err))
		} else {
			isRm = true
		}
	}
	defer func() { pid.isRun = false }()
	if pid.cfg.Debug() && pid.isRun {
		switch isRm {
		case true:
			pid.log().Infof(tplPidEndFileDeleted, pid.pfnm)
		default:
			pid.log().Info(tplPidEnd)
		}
	}

	return
}

func (pid *impl) isSkip() (ret bool) {
	const cmdVersion, cmdConfig = `version`, `config`

	// Для стандартной команды версии приложения миграцию не запускаем.
	switch {
	case pid.cfg.Command() == cmdVersion:
		ret = true
	case strings.HasPrefix(pid.cfg.Command(), cmdConfig):
		ret = true
	}

	return
}
