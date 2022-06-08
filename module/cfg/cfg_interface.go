// Package cfg
package cfg

import (
	"os/user"

	kitModuleBus "github.com/webnice/kit/module/bus"
	kitModuleLogLevel "github.com/webnice/kit/module/log/level"
	kitModuleUuid "github.com/webnice/kit/module/uuid"
	kitTypes "github.com/webnice/kit/types"
)

// Interface Интерфейс пакета.
type Interface interface {
	// ИНТЕРФЕЙСЫ.

	// Bus Интерфейс шины данных приложения.
	Bus() kitModuleBus.Interface

	// Gist Интерфейс к служебным методам конфигурации приложения.
	Gist() Essence

	// Version Интерфейс к методам получения версии приложения.
	Version() Version

	// UUID Интерфейс к методам генерации и работы с UUID идентификаторами.
	UUID() kitModuleUuid.Interface

	// RawWriter Интерфейс вывода потоковых сообщений.
	RawWriter() kitTypes.SyncWriter

	// МЕНЕДЖЕР ЛОГИРОВАНИЯ.

	// Log Интерфейс к методам логирования.
	Log() kitTypes.Logger

	// Loglevel Текущий уровень логирования приложения.
	Loglevel() kitModuleLogLevel.Level

	// МЕТОДЫ.

	// AppName Название приложения.
	AppName() string

	// Command Функция возвращает команду с которой было запущено приложение.
	Command() (ret string)

	// CommandFull Функция возвращает команду с которой было запущено приложение.
	CommandFull() []string

	// Debug Функция возвращает текуще значение режима отладки приложения.
	Debug() (ret bool)

	// ForkWorker Параметры работы в режиме fork worker.
	ForkWorker() *kitTypes.BootstrapConfigurationForkWorker

	// IsForkWorker Возвращает истину для режима запуска приложения в качестве подчинённого процесса основного приложения.
	// В этом режиме приложению нельзя использовать STDIN/STDOUT/STDERR, эти потоки используются для обмена данными
	// с управляющим приложением.
	// Так же, этот режим устанавливается в случае запуска приложения как daemon процесса, в этом режиме запуска, тоже
	// нельзя использовать STDIN/STDOUT/STDERR, так как режим daemon, для приложения, возможен только после полного
	// закрытия этих потоков.
	IsForkWorker() (ret bool)

	// Runlevel Возвращает текущее значение уровня работы приложения.
	Runlevel() uint16

	// RunlevelMap Возвращает карту, описывающую план переключения уровней работы приложения.
	RunlevelMap() []uint16

	// RunlevelSubscribe Подписка на события изменения уровня работы приложения.
	RunlevelSubscribe(fn RunlevelSubscriberFn) (err error)

	// RunlevelUnsubscribe Отписка от событий изменения уровня работы приложения.
	RunlevelUnsubscribe(fn RunlevelSubscriberFn) (err error)

	// RunlevelAutoincrement Режим работы автоматического увеличения уровня работы приложения.
	RunlevelAutoincrement() (ret bool)

	// Targetlevel Возвращает текущее значение целевого уровня работы приложения.
	Targetlevel() uint16

	// User Текущий пользователь операционной системы.
	User() *user.User

	// AbsolutePath Функция преобразует относительный путь в абсолютный путь к директории или файлу.
	// Учитывается символ '~' обозначающий домашнюю директорию текущего пользователя.
	AbsolutePath(pth string) (ret string)

	// DirectoryHome Значение домашней директории приложения.
	DirectoryHome() string

	// DirectoryWorking Значение рабочей директории приложения.
	DirectoryWorking() string

	// DirectoryWorkingChdir Функция выполняет переход в рабочую директорию приложения.
	// Возвращается ошибка операционной системы, с причиной не возможности перейти в рабочую директорию.
	DirectoryWorkingChdir() (err error)

	// DirectoryTemp Значение директории для временных файлов.
	DirectoryTemp() string

	// DirectoryCache Значение директории для файлов кеша.
	DirectoryCache() string

	// DirectoryConfig Значение директории для подключаемых или дополнительных конфигураций приложения.
	DirectoryConfig() string

	// FileConfig Значение пути и имени конфигурационного файла приложения.
	FileConfig() string

	// FilePid Значение пути и имени PID файла приложения.
	FilePid() string

	// FileState Значение пути и имени файла хранения состояния приложения.
	FileState() string

	// FileSocket Значение пути и имени сокет файла коммуникаций с приложением.
	FileSocket() string

	// ConfigurationUnionSprintf Печать объединённой конфигурации приложения в строку.
	ConfigurationUnionSprintf() (ret string)

	// ОШИБКИ

	// Errors Все ошибки известного состояния, которые может вернуть приложение или функция.
	Errors() *Error
}
