package types

import (
	"regexp"
	"time"
)

// Component Интерфейс компоненты приложения
type Component interface {
	// Preferences Настройки компоненты приложения.
	// Опрашивается один раз, при Runlevel=2
	Preferences() (ret ComponentPreferences)

	// Initiate Запуск инициализации компоненты приложения и подготовки к выполнению основного кода компоненты.
	// В этой функции компонента должен проверить всё что ей требуется для работы и выполнить подготовку к работе,
	// функция не должна блокировать работу приложения на долгое время, во избежание завершения приложения с ошибкой
	// инициализации компоненты приложения.
	// Выполнение функции Initiate() ограничено по времени, указывается в InitiateTimeout
	// Функция вызывается один раз при Runlevel=8
	Initiate() (err error)

	// Do Выполнение компоненты приложения
	// Функция возвращает:
	// - (levelDone) - Флаг остановки автоматического переключения уровней работы приложения.
	// - (levelExit) - Флаг переключения уровня работы приложения на уровень завершения работы приложения.
	// - (err)       - Объект с ошибкой выполнения компоненты приложения. Если хотя бы одна компонента приложения
	//                 вернёт ошибку, приложение начнёт завершение работы с ошибкой выполнения.
	// Выполнение функции Do() никак не ограничено по времени.
	// Функция вызывается один раз, начиная с Runlevel=10 и заканчивая Runlevel=65534.
	Do() (levelDone bool, levelExit bool, err error)

	// Finalize Функция вызывается перед завершением компоненты и приложения в целом.
	// Каждый компонент, при вызове этой функции должен остановить все запущенные процессы, закрыть файлы,
	// сохранить все данные и завершить работу в штатном режиме.
	// Выполнение функции никак не ограничено по времени, но не рекомендуется долго блокировать приложение на этой
	// стадии, так как скорее всего приложение будет "убито" операционной системой, если не будет отвечать.
	// Функция вызывается один раз, Runlevel=65535
	Finalize() (err error)
}

// ComponentInfo Данные зарегистрированного компонента
type ComponentInfo struct {
	Before          []*regexp.Regexp // Массив приоритета запуска "ДО"
	After           []*regexp.Regexp // Массив приоритета запуска "ПОСЛЕ"
	Require         []*regexp.Regexp // Массив строгих зависимостей "ТРЕБУЕТ"
	Conflict        []*regexp.Regexp // Массив строгих правил "КОНФЛИКТУЕТ"
	Command         []string         // Массив команд
	InitiateTimeout time.Duration    // Максимальное время ожидания выполнения Initiate()
	Runlevel        uint16           // Минимальный уровень приложения, при котором выполняется запуск компоненты
	Component       Component        // Интерфейс зарегистрированного компонента
	ComponentName   string           // Название зарегистрированного компонента
	IsDisable       bool             // Флаг активности компоненты, для значения true, компонента отключена
	IsInitiate      bool             // Флаг указывающий была ли выполнена функция Initiate()
	IsDo            bool             // Флаг указывающий была ли выполнена функция Do()
	IsFinalize      bool             // Флаг указывающий была ли выполнена функция Finalize()
}

// ComponentPreferences Конфигурация компоненты приложения
type ComponentPreferences struct {
	// InitiateTimeout Максимальное время ожидания выполнения Initiate() при запуске системы
	// Эта опция переопределяет время ожидания по умолчанию, чтобы гарантировать правильный запуск приложения.
	// Если указано 0, тогда используется значение таймаута по умолчанию.
	// Значение по умолчанию: 1m
	InitiateTimeout time.Duration

	// Runlevel Минимальный уровень работы приложения, при котором выполняется запуск компоненты, функция Do()
	// Значение по умолчанию 0 - не учитывается, компонент стартует на уровне Targetlevel.
	// При работе приложения, уровни проходят значения от 0 до 65535, где:
	// 0           - начало работы приложения.
	// Targetlevel - уровень запуска всех компонентов с Runlevel=0 и ожидание (ожидание зависит от значения Targetlevel)
	// 65535       - завершение работы приложения.
	// Условия запуска функции Do() каждого компонента следующие:
	// Если Runlevel=0, компонента будет вызвана при достижении приложением Runlevel=Targetlevel.
	// Если Runlevel<10 или Runlevel>65534, приложение завершится с ошибкой настроек компоненты.
	// Если Runlevel>=10 и Runlevel<=Targetlevel, компонента будет вызвана при достижении Targetlevel.
	// Если Runlevel>Targetlevel, компонента будет вызвана перед завершением приложения,
	//                            при достижении Runlevel=65535
	// Значение по умолчанию: 0
	Runlevel uint16

	// IsDisable Для значения true, компонента отключена, но зарегистрирована и доступна.
	// Заставляет полностью игнорировать компоненту, при этом, регистрация компоненты остаётся, а последним
	// вызовом компоненты является функция Preferences().
	// Так же, в этом режиме не проверяются зависимости компоненты и не вызывается функция инициализации и завершения.
	IsDisable bool

	// Before Массив приоритета запуска "ДО"
	// Массив состоит из regexp правил поиска по названиям пакетов, перед которыми должна выполняться компонента.
	// Сортировка "ДО" может быть перекрыта более приоритетной сортировкой "ПОСЛЕ".
	// Если перечисленных компонентов в приложении не зарегистрировано, тогда сортировка запуска не выполняется.
	// Если массив пустой, тогда компонента выполняется в порядке регистрации или в соответствии с другими правилами.
	Before []string

	// After Массив приоритета запуска "ПОСЛЕ"
	// Массив состоит из regexp правил поиска по названиям пакетов, после которых должна выполняться компонента.
	// Сортировка "ПОСЛЕ" имеет приоритет над сортировкой "ДО".
	// Если массив пустой, тогда компонента выполняется в порядке регистрации или в соответствии с другими правилами.
	After []string

	// Require Массив строгих зависимостей "ТРЕБУЕТ"
	// Массив состоит из regexp правил поиска по названиям пакетов, все перечисленные в "ТРЕБУЕТ" компоненты,
	// должны быть зарегистрированы в приложении.
	// Компонента с зависимостями "ТРЕБУЕТ", запускается в соответствии с правилами Before() и After(), то есть,
	// зависимость "ТРЕБУЕТ" проверяет обязательное наличие пакета, но не влияет на порядок запуска компонентов.
	// Если в приложении не зарегистрирована хотя бы одна из указанный компонент, приложение завершится с
	// ошибкой исключения - "отсутствие зависимости", исключение будет вызвано до начала выполнения функций Initiate().
	// Если массив пустой, тогда приложение не проверяет зависимости для компоненты.
	Require []string

	// Conflict Массив строгих правил "КОНФЛИКТУЕТ"
	// Массив состоит из regexp правил поиска по названиям пакетов, ни один из пакетов подпадающих под
	// правило "КОНФЛИКТУЕТ" не должен быть зарегистрирован в приложении в качестве компоненты приложения.
	// Если в приложении зарегистрирован хотя бы один из перечисленный пакетов, приложение завершится с ошибкой
	// исключения - "конфликт зависимости", исключение будет вызвано до начала выполнения функций Initiate().
	// Если массив пустой, тогда приложение не проверяет конфликтующие зависимости для компоненты.
	Conflict []string

	// Command Массив команд, при указании которых, выполняется запуск компоненты
	// Массив состоит из чувствительных к регистру строк в кодировке UTF-8.
	// Команды указывают условия запуска функции Do() для аргумента командной строки приложения.
	// Если список команд пустой, тогда компонента запускается для любой команды.
	// Функции инициализации и финализации компоненты запускаются всегда, без учёта команд.
	Command []ComponentCommand

	// Flag Массив флагов, при указании, в параметры приложения добавляется возможность загрузки данных из флагов
	// командной строки. Флаги, описанные в данном объекте, запрашиваются глобально, вне зависимости от
	// указания команды. Помимо глобальных флагов, флаги можно указать так же у каждой команды приложения.
	Flag []ComponentFlag
}

// ComponentCommand Описание структуры команды приложения.
type ComponentCommand struct {
	GroupKey         string      // Ключ группы команд, если пустой, группа не создаётся.
	GroupTitle       string      // Заголовок группы команд, может быть пустым.
	GroupDescription string      // Описание группы команд. Если пустое, тогда группа не создаётся.
	Command          string      // Название команды. Пустые команды не создаются.
	Description      string      // Описание команды, отображается в помощи пользователю.
	IsDefault        bool        // Команда по умолчанию.
	IsHidden         bool        // Команда скрыта, не отображается в помощи.
	Value            interface{} // Ссылка на структуру значений с мета информацией, в неё же будут загружены указанное в CLI или ENV, значения.
}

// ComponentFlag Описание структуры флагов компоненты приложения.
type ComponentFlag struct {
	ShortKey    rune        // Короткий, односимвольный синоним флага, может быть пустым.
	Flag        string      // Полное наименование флага, если пустой, флаг игнорируется.
	Description string      // Помощь для пользователя, описывающая назначение флага.
	Environment string      // Наименование переменной окружения из которой может быть взято значение флага.
	Placeholder string      // Значение флага, которое будет указано при отображении в помощи.
	IsRequired  bool        // Обязательный флаг.
	IsHidden    bool        // Флаг скрыт, не отображается в помощи.
	Value       interface{} // Ссылка на переменную, в которую будет загружено указанное в CLI или ENV, значение.
}
