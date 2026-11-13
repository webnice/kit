package cfg

import (
	"math"
	"strings"

	kitModuleCfgCli "github.com/webnice/kit/v4/module/cfg/cli"
	kitModuleLog "github.com/webnice/kit/v4/module/log"
	kitModuleLogLevel "github.com/webnice/kit/v4/module/log/level"
	kitTypes "github.com/webnice/kit/v4/types"

	"github.com/Masterminds/semver/v3"
)

// Создание объекта и возвращение интерфейса Essence.
func newEssence(parent *impl) Essence {
	var essence = &gist{parent: parent}
	return essence
}

// App Точка запуска приложения.
func (essence *gist) App() {
	defer func() {
		if len(essence.parent.error) == 0 {
			return
		}
		essence.exitWithCode()
	}()
	if essence.parent.main.Fn == nil {
		essence.parent.error = append(essence.parent.error, essence.parent.Errors().ApplicationMainFuncNotFound.Bind())
		return
	}
	essence.safeLaunchApplication()
}

// Cfg Интерфейс конфигурации.
func (essence *gist) Cfg() Interface { return essence.parent }

// Log Интерфейс к методам логирования.
func (essence *gist) Log() kitTypes.Logger { return essence.parent.Log() }

// Logger Интерфейс менеджера логирования.
func (essence *gist) Logger() kitModuleLog.Logger { return essence.parent.logger }

// CLI Интерфейс к методам работы с параметрами командной строки и переменных окружения.
func (essence *gist) CLI() kitModuleCfgCli.Interface { return essence.parent.cli }

// AppName Установка значения названия приложения.
func (essence *gist) AppName(name string) Essence {
	essence.parent.bootstrapConfiguration.ApplicationName = name
	return essence
}

// Version Присвоение версии и номера сборки приложения в семантике "Semantic Versioning 2.0.0".
// Согласно документации https://semver.org/
func (essence *gist) Version(version string, build string) Essence {
	const keyBuild = "+build."
	var (
		tmp string
		err error
	)

	if tmp = version; build != "" {
		tmp = version + keyBuild + build
	}
	if essence.parent.main.Version, err = semver.NewVersion(tmp); err != nil {
		essence.parent.error = append(essence.parent.error, essence.parent.Errors().ApplicationVersion.Bind(err))
		return essence
	}

	return essence
}

// Debug Присвоение нового значения режима отладки приложения.
func (essence *gist) Debug(debug bool) Essence {
	essence.parent.bootstrapConfiguration.ApplicationDebug = debug
	essence.parent.logger.Debug(debug)
	essence.parent.bus.Gist().Debug(debug)
	essence.parent.srv.Gist().Debug(debug)
	essence.parent.ans.Gist().Debug(debug)

	return essence
}

// NoPidFileSet Установка значения флага запрета работы плагина pidfile.
// Используется для режима работы приложения в качестве утилиты командной строки.
func (essence *gist) NoPidFileSet(noPidFile bool) Essence {
	essence.parent.noPidFile = noPidFile

	return essence
}

// NoPidFileGet Получение значения флага запрета работы плагина pidfile.
func (essence *gist) NoPidFileGet() bool { return essence.parent.noPidFile }

// NoDbMigrationSet Установка значения флага запрета применения миграций базы данных.
// Используется для режима работы приложения в качестве утилиты командной строки.
func (essence *gist) NoDbMigrationSet(noDbMigration bool) Essence {
	essence.parent.noDbMigration = noDbMigration

	return essence
}

// NoDbMigrationGet Получение значения флага запрета применения миграций базы данных.
func (essence *gist) NoDbMigrationGet() bool { return essence.parent.noDbMigration }

// UpdateBootstrapConfigurationPathValue Коррекция значений стартовой конфигурация приложения, если они были загружены
// из командной строки или переменной окружения, то есть не равны значениям по умолчанию.
func (essence *gist) UpdateBootstrapConfigurationPathValue() Essence {
	// Коррекция значений, если они были загружены из командной строки или переменной окружения
	if essence.parent.bootstrapConfiguration.HomeDirectory != defaultHomeDirectory() {
		essence.parent.AbsolutePathAndUpdate(&essence.parent.bootstrapConfiguration.HomeDirectory)
	}
	if essence.parent.bootstrapConfiguration.WorkingDirectory != defaultWorkingDirectory() {
		essence.parent.AbsolutePathAndUpdate(&essence.parent.bootstrapConfiguration.WorkingDirectory)
	}
	if essence.parent.bootstrapConfiguration.TempDirectory != defaultTempDirectory() {
		essence.parent.AbsolutePathAndUpdate(&essence.parent.bootstrapConfiguration.TempDirectory)
	}
	if essence.parent.bootstrapConfiguration.CacheDirectory != defaultCacheDirectory() {
		essence.parent.AbsolutePathAndUpdate(&essence.parent.bootstrapConfiguration.CacheDirectory)
	}
	if essence.parent.bootstrapConfiguration.ConfigDirectory != defaultConfigDirectory() {
		essence.parent.AbsolutePathAndUpdate(&essence.parent.bootstrapConfiguration.ConfigDirectory)
	}
	if essence.parent.bootstrapConfiguration.LogDirectory != defaultLogDirectory() {
		essence.parent.AbsolutePathAndUpdate(&essence.parent.bootstrapConfiguration.LogDirectory)
	}

	return essence
}

// CommandFull Присвоение значения полной команды с которой было запущено приложение.
func (essence *gist) CommandFull(cmd []string) Essence {
	essence.parent.command = make([]string, 0, len(cmd))
	essence.parent.command = append(essence.parent.command, cmd...)
	return essence
}

// ForkWorkerMode Присвоение нового значения режима запуска в качестве подчинённого процесса основного приложения.
func (essence *gist) ForkWorkerMode(isForkWorker bool) Essence {
	essence.parent.isForkWorker = isForkWorker
	return essence
}

// ErrorAppend Добавление ошибки в стек ошибок приложения.
//
// Операция потокобезопасна и защищена errorMutex. Метод можно вызывать из runlevel-подписчиков,
// компонентов и служебных goroutine; ошибки публикуются в общий список и выводятся при завершении приложения.
func (essence *gist) ErrorAppend(err error) Essence {
	essence.parent.errorMutex.Lock()
	essence.parent.error = append(essence.parent.error, err)
	essence.parent.errorMutex.Unlock()
	return essence
}

// IsError Возвращает истину, если есть ошибки приложения.
//
// Чтение списка ошибок защищено errorMutex и безопасно для конкурентного ErrorAppend.
func (essence *gist) IsError() bool {
	essence.parent.errorMutex.RLock()
	defer essence.parent.errorMutex.RUnlock()
	return len(essence.parent.error) > 0
}

// Runlevel Установка или увеличение значения уровня работы приложения.
//
// Функция отправляет запрос в runlevel-диспетчер и блокируется до подтверждения применения нового уровня.
// Поведение функции в зависимости от значения аргумента runlevel:
//   - runlevel>0: выполняется установка нового значения; новое значение не может быть меньше текущего;
//   - runlevel=0: значение Runlevel увеличивается на единицу относительно текущего уровня.
//
// Попытка установить уровень меньше текущего добавляет ошибку RunlevelCantLessCurrentLevel и не меняет уровень.
func (essence *gist) Runlevel(runlevel uint16) Essence {
	var (
		current uint16
		msg     *runLevelUp
	)

	current = essence.parent.Runlevel()
	if runlevel > 0 && current > runlevel {
		essence.ErrorAppend(
			essence.parent.Errors().RunlevelCantLessCurrentLevel.Bind(current, runlevel),
		)
		return essence
	}
	if runlevel == 0 {
		runlevel = current + 1
	}
	// Отправка сообщения об изменении уровня работы приложения
	msg = &runLevelUp{
		newLever: runlevel,
		done:     make(chan struct{}),
	}
	essence.parent.runLevelChangeChan <- msg
	// Ожидание применения нового значения и вызова всех подписчиков на событие изменения уровня работы приложения
	<-msg.done
	// Закрытие канала обратной связи
	close(msg.done)

	return essence
}

// RunlevelNext Переключение значения уровня работы приложения на следующий уровень.
// От 01 до 10 переключается на единицу, последовательно.
// От 10 до 65535 переключается согласно карте уровней, если текущий уровень не входит в карту - последовательно.
// На уровне 65535 переключение останавливается, вызов функции не меняет уровень.
func (essence *gist) RunlevelNext() Essence {
	var (
		current  uint16
		mapLevel []uint16
		n, next  int
	)

	current = essence.parent.Runlevel()
	mapLevel = essence.parent.RunlevelMap()
	// Если достигли максимального значения, тогда выход
	if current == math.MaxUint16 {
		return essence
	}
	// Поиск текущего уровня в карте переключения уровней
	for n = range mapLevel {
		if mapLevel[n] == current {
			next = n + 1
			break
		}
	}
	// Если не найден текущий уровень в карте переключения уровней, тогда следующий берётся не из карты, увеличением +1
	if next > len(mapLevel)-1 {
		next = 0
	}
	// Для всех уровней, не входящих в карту переключения уровней, увеличение последовательно на единицу
	if next == 0 {
		return essence.Runlevel(0)
	}

	return essence.Runlevel(mapLevel[next])
}

// RunlevelNextAsync Асинхронное переключение значения уровня работы приложения на следующий уровень.
func (essence *gist) RunlevelNextAsync() { go essence.RunlevelNext() }

// RunlevelDefault Уровень выполнения компонентов приложения по умолчанию.
func (essence *gist) RunlevelDefault() uint16 { return defaultRunlevel }

// RunlevelAutoincrementStop Отключение режима автоматического увеличения уровня работы приложения.
func (essence *gist) RunlevelAutoincrementStop() Essence {
	essence.parent.runLevelMutex.Lock()
	essence.parent.runLevelStopAutoincrement = true
	essence.parent.runLevelMutex.Unlock()
	return essence
}

// RunlevelAutoincrementStart Включение режима автоматического увеличения уровня работы приложения.
func (essence *gist) RunlevelAutoincrementStart() Essence {
	essence.parent.runLevelMutex.Lock()
	essence.parent.runLevelStopAutoincrement = false
	essence.parent.runLevelMutex.Unlock()
	return essence
}

// RunlevelExit Переключение значения уровня работы приложения на уровень завершения работы приложения.
func (essence *gist) RunlevelExit() Essence { return essence.Runlevel(defaultRunlevelExit) }

// RunlevelExitAsync Переключение значения уровня работы приложения на уровень завершения работы приложения.
func (essence *gist) RunlevelExitAsync() { go essence.RunlevelExit() }

// Targetlevel Установка значения целевого уровня работы приложения.
func (essence *gist) Targetlevel(atl uint16) Essence {
	essence.parent.bootstrapConfiguration.ApplicationTargetlevel = atl
	return essence
}

// Атомарная последовательная регистрация разных объектов приложения.
func (essence *gist) syncRegistration(ktc kitTypes.Component, name string) {
	var componentInfo *kitTypes.ComponentInfo

	essence.parent.comp.ComponentMutex.Lock()
	componentInfo = &kitTypes.ComponentInfo{
		InitiateTimeout: defaultInitiateTimeout(),
		Component:       ktc,
		ComponentName:   name,
	}
	essence.parent.comp.Registered = append(essence.parent.comp.Registered, componentInfo)
	essence.parent.comp.ComponentMutex.Unlock()
}

// Registration Регистрация разных объектов приложения.
func (essence *gist) Registration(name string, obj any) (err error) {
	var (
		ktc kitTypes.Component
		ok  bool
	)

	switch obj.(type) {
	case kitTypes.Component:
		if ktc, ok = obj.(kitTypes.Component); ok {
			essence.syncRegistration(ktc, name)
		}
	default:
		err = essence.Cfg().Errors().ApplicationRegistrationUnknownObject.Bind(name)
	}

	return
}

// DirectoryHome Установка значения домашней директория приложения.
func (essence *gist) DirectoryHome(ahd string) Essence {
	essence.parent.bootstrapConfiguration.HomeDirectory = strings.TrimSpace(ahd)
	return essence
}

// DirectoryWorking Установка значения рабочей директории приложения.
func (essence *gist) DirectoryWorking(awd string) Essence {
	essence.parent.bootstrapConfiguration.WorkingDirectory = strings.TrimSpace(awd)
	return essence
}

// DirectoryTemp Установка значения директории для временных файлов.
func (essence *gist) DirectoryTemp(atd string) Essence {
	essence.parent.bootstrapConfiguration.TempDirectory = strings.TrimSpace(atd)
	return essence
}

// DirectoryCache Установка значения директории для файлов кеша.
func (essence *gist) DirectoryCache(acd string) Essence {
	essence.parent.bootstrapConfiguration.CacheDirectory = strings.TrimSpace(acd)
	return essence
}

// DirectoryConfig Установка значения директории для подключаемых или дополнительных конфигураций приложения.
func (essence *gist) DirectoryConfig(afd string) Essence {
	essence.parent.bootstrapConfiguration.ConfigDirectory = strings.TrimSpace(afd)
	return essence
}

// DirectoryLog Установка значения директории для файлов журнала приложения.
func (essence *gist) DirectoryLog(ald string) Essence {
	essence.parent.bootstrapConfiguration.LogDirectory = strings.TrimSpace(ald)
	return essence
}

// FileConfig Установка значения пути и имени конфигурационного файла приложения.
func (essence *gist) FileConfig(cfn string) Essence {
	essence.parent.bootstrapConfiguration.ConfigFile = strings.TrimSpace(cfn)
	return essence
}

// FilePid Установка значения пути и имени PID файла приложения.
func (essence *gist) FilePid(pfn string) Essence {
	essence.parent.bootstrapConfiguration.PidFile = strings.TrimSpace(pfn)
	return essence
}

// FileState Установка значения пути и имени файла хранения состояния приложения.
func (essence *gist) FileState(stf string) Essence {
	essence.parent.bootstrapConfiguration.StateFile = strings.TrimSpace(stf)
	return essence
}

// FileSocket Установка значения пути и имени сокет файла коммуникаций с приложением.
func (essence *gist) FileSocket(sof string) Essence {
	essence.parent.bootstrapConfiguration.SocketFile = strings.TrimSpace(sof)
	return essence
}

// LogLevel Установка значения уровня логирования приложения.
func (essence *gist) LogLevel(ll kitModuleLogLevel.Level) Essence {
	essence.parent.bootstrapConfiguration.LogLevel = ll
	return essence
}

// AbsolutePathAndUpdate Функция преобразует относительный путь в абсолютный путь,
// проверяет равно ли новое значение старому, если значение изменилось, тогда обновляет значение в указанной
// переменной и возвращает результирующее значение.
func (essence *gist) AbsolutePathAndUpdate(dir *string) (ret string) {
	return essence.parent.AbsolutePathAndUpdate(dir)
}
