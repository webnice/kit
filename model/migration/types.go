package migration

import kitModuleDbSql "github.com/webnice/kit/v4/module/db/sql"

// Interface Интерфейс пакета.
type Interface interface {
	// CurrentVersion Возвращается текущая версия схемы базы данных.
	CurrentVersion() (ver *DbVersion, err error)

	// ОШИБКИ

	// Errors Справочник всех ошибок пакета.
	Errors() *Error
}

// Объект сущности пакета.
type impl struct {
	kitModuleDbSql.Implementation
}
