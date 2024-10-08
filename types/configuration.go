package types

// ConfigurationDefaulter Интерфейс предназначен для заполнения структуры конфигурации значениями по умолчанию.
// Структура, реализующая данный интерфейс, заполняется значениями по умолчанию через вызов функции Default().
// При этом, значения прописанные в тегах структуры "default-value", игнорироваться.
type ConfigurationDefaulter interface {
	// Default Функция установки значений по умолчанию для структуры конфигурации.
	Default() (err error)
}
