// Package cfg
package cfg

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/webnice/debug"

	kitModuleBus "github.com/webnice/kit/v3/module/bus"
	kitModuleLogLevel "github.com/webnice/kit/v3/module/log/level"
	kitModuleUuid "github.com/webnice/kit/v3/module/uuid"
	kitTypes "github.com/webnice/kit/v3/types"
)

// Get Функция возвращает интерфейс объекта пакета.
func Get() Interface { return singleton }

// RegistrationMain Регистрация основной функции приложения.
func RegistrationMain(fn kitTypes.MainFn) Interface {
	var fnName string

	singleton.main.FnMutex.Lock()
	if singleton.main.Fn != nil {
		fnName = getFuncFullName(RegistrationMain)
		singleton.error = append(singleton.error, Errors().ApplicationMainFuncAlreadyRegistered(0))
		panic(fnName + "\n" + Errors().ApplicationMainFuncAlreadyRegistered(0).Error())
		return singleton
	}
	singleton.main.Fn = fn
	singleton.main.FnMutex.Unlock()

	return singleton
}

// Bus Интерфейс шины данных приложения.
func (cfg *impl) Bus() kitModuleBus.Interface { return cfg.bus }

// Gist Интерфейс к служебным методам конфигурации приложения.
func (cfg *impl) Gist() Essence { return cfg.essence }

// Version Интерфейс к методам получения версии приложения.
func (cfg *impl) Version() Version { return cfg.version }

// UUID Интерфейс к методам генерации и работы с UUID идентификаторами.
func (cfg *impl) UUID() kitModuleUuid.Interface { return cfg.uuid }

// RawWriter Интерфейс вывода потоковых сообщений.
func (cfg *impl) RawWriter() kitTypes.SyncWriter { return cfg.rawWriter }

// Log Интерфейс к методам логирования.
func (cfg *impl) Log() kitTypes.Logger { return cfg.rec }

// Loglevel Значение уровня логирования приложения.
func (cfg *impl) Loglevel() kitModuleLogLevel.Level {
	return kitModuleLogLevel.ParseLevelInt64(int64(cfg.bootstrapConfiguration.LogLevel))
}

// AppName Название приложения.
func (cfg *impl) AppName() string { return cfg.bootstrapConfiguration.ApplicationName }

// Command Функция возвращает команду с которой было запущено приложение.
func (cfg *impl) Command() (ret string) {
	if len(cfg.command) > 0 {
		ret = cfg.command[0]
	}

	return
}

// CommandFull Функция возвращает команду с которой было запущено приложение.
func (cfg *impl) CommandFull() []string { return cfg.command }

// Debug Функция возвращает текуще значение режима отладки приложения.
func (cfg *impl) Debug() (ret bool) { return cfg.bootstrapConfiguration.ApplicationDebug }

// ForkWorker Параметры работы в режиме fork worker.
func (cfg *impl) ForkWorker() *kitTypes.BootstrapConfigurationForkWorker {
	if cfg.cacheForkWorker == nil {
		cfg.cacheForkWorker = &kitTypes.BootstrapConfigurationForkWorker{
			Master:    cfg.bootstrapConfiguration.ApplicationForkWorkerMaster,
			Component: cfg.bootstrapConfiguration.ApplicationForkWorkerComponent,
			Target:    cfg.bootstrapConfiguration.ApplicationForkWorkerTarget,
		}
	}

	return cfg.cacheForkWorker
}

// IsForkWorker Возвращает истину для режима запуска приложения в качестве подчинённого процесса основного приложения.
func (cfg *impl) IsForkWorker() (ret bool) { return cfg.isForkWorker }

// Errors Все ошибки известного состояния, которые может вернуть приложение или функция.
func (cfg *impl) Errors() *Error { return Errors() }

// Targetlevel Возвращает текущее значение целевого уровня работы приложения.
func (cfg *impl) Targetlevel() uint16 { return cfg.bootstrapConfiguration.ApplicationTargetlevel }

// User Текущий пользователь операционной системы.
func (cfg *impl) User() *user.User { return cfg.user }

// AbsolutePath Функция преобразует относительный путь в абсолютный путь к директории или файлу.
// Учитывается символ '~' обозначающий домашнюю директорию текущего пользователя.
func (cfg *impl) AbsolutePath(pth string) (ret string) { return AbsolutePath(pth) }

// AbsolutePathAndUpdate Функция преобразует относительный путь в абсолютный путь,
// проверяет равно ли новое значение старому, если значение изменилось, тогда обновляет значение в указанной
// переменной и возвращает результирующее значение.
func (cfg *impl) AbsolutePathAndUpdate(dir *string) (ret string) {
	if dir == nil {
		return
	}
	if ret = cfg.AbsolutePath(*dir); *dir != ret {
		*dir = ret
	}

	return
}

// DirectoryHome Значение домашней директории приложения.
func (cfg *impl) DirectoryHome() string {
	return cfg.AbsolutePathAndUpdate(&cfg.bootstrapConfiguration.HomeDirectory)
}

// DirectoryWorking Значение рабочей директории приложения.
func (cfg *impl) DirectoryWorking() string {
	return cfg.AbsolutePathAndUpdate(&cfg.bootstrapConfiguration.WorkingDirectory)
}

// DirectoryWorkingChdir Функция выполняет переход в рабочую директорию приложения.
// Возвращается ошибка операционной системы, с причиной не возможности перейти в рабочую директорию.
func (cfg *impl) DirectoryWorkingChdir() (err error) {
	var (
		awd string
		cwd string
	)

	awd = cfg.DirectoryWorking()
	cwd, _ = os.Getwd()
	// Рабочая директория из конфигурации совпадает с текущей рабочей директорией, смена директории не требуется.
	if strings.EqualFold(cwd, awd) {
		return
	}
	// Системный вызов смены рабочей директории.
	if err = os.Chdir(awd); err != nil {
		err = fmt.Errorf("syscall chdir(%q) error: %s", awd, err)
		return
	}

	return
}

// DirectoryTemp Значение директории для временных файлов.
func (cfg *impl) DirectoryTemp() string {
	return cfg.AbsolutePathAndUpdate(&cfg.bootstrapConfiguration.TempDirectory)
}

// DirectoryCache Значение директории для файлов кеша.
func (cfg *impl) DirectoryCache() string {
	return cfg.AbsolutePathAndUpdate(&cfg.bootstrapConfiguration.CacheDirectory)
}

// DirectoryConfig Значение директории для подключаемых или дополнительных конфигураций приложения.
func (cfg *impl) DirectoryConfig() string {
	return cfg.AbsolutePathAndUpdate(&cfg.bootstrapConfiguration.ConfigDirectory)
}

// FileConfig Значение пути и имени конфигурационного файла приложения.
func (cfg *impl) FileConfig() string {
	return cfg.AbsolutePathAndUpdate(&cfg.bootstrapConfiguration.ConfigFile)
}

// FilePid Значение пути и имени PID файла приложения.
func (cfg *impl) FilePid() string { return cfg.bootstrapConfiguration.PidFile }

// FileState Значение пути и имени файла хранения состояния приложения.
func (cfg *impl) FileState() string { return cfg.bootstrapConfiguration.StateFile }

// FileSocket Значение пути и имени сокет файла коммуникаций с приложением.
func (cfg *impl) FileSocket() string { return cfg.bootstrapConfiguration.SocketFile }

// ConfigurationUnionSprintf Печать объединённой конфигурации приложения в строку.
func (cfg *impl) ConfigurationUnionSprintf() (ret string) {
	const structName = `*struct.UnionConfiguration`
	ret = rexUnionStructureHeader.ReplaceAllString(debug.DumperString(cfg.conf.Union), structName)
	return
}
