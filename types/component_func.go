package types

// MainFn Функция-точка запуска и завершения приложения.
type MainFn func() (code uint8, err error)

// ComponentPreferencesFn Тип функции загрузки и обработки настроек компонента.
type ComponentPreferencesFn func(*ComponentInfo) (*ComponentInfo, ErrorWithCode)

// ComponentConflictFn Тип функции проверки конфликтов между компонентами.
type ComponentConflictFn func([]*ComponentInfo) ErrorWithCode

// ComponentRequiresFn Тип функции проверки зависимостей между компонентами.
type ComponentRequiresFn func([]*ComponentInfo) ErrorWithCode

// ComponentSortFn Тип функции сортировки компонентов в соответствии с настройками (before) и (after).
type ComponentSortFn func([]*ComponentInfo) ErrorWithCode

// ComponentInitiateFn Тип функции вызова функции Initiate компонента с контролем длительности выполнения.
type ComponentInitiateFn func(*ComponentInfo) ErrorWithCode

// ComponentDoFn Тип функции вызова функции Do компонента.
type ComponentDoFn func(*ComponentInfo) ErrorWithCode

// ComponentFinalizeFn Тип функции вызова функции Finalize компонента.
type ComponentFinalizeFn func(*ComponentInfo) ErrorWithCode
