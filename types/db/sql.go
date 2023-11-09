package db

import kitModuleDbSqlTypes "github.com/webnice/kit/v4/module/db/sql/types"

// DatabaseSqlConfiguration Конфигурация подключения к реляционной базе данных SQL.
type DatabaseSqlConfiguration struct {
	SqlDB kitModuleDbSqlTypes.Configuration `yaml:"SqlDB"`
}
