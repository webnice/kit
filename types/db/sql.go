// Package db
package db

import kitModuleDbSqlTypes "github.com/webnice/kit/v3/module/db/sql/types"

// DatabaseSqlConfiguration Конфигурация подключения к базе данных SQL.
type DatabaseSqlConfiguration struct {
	SqlDB kitModuleDbSqlTypes.Configuration `yaml:"SqlDB"`
}
