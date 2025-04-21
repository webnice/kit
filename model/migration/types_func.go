package migration

//go:generate db2struct create "" "goose_db_version" "migration" "DbVersion" "types_model.go"

// TableName Явное указание названия таблицы для ORM gorm.
func (cpt *DbVersion) TableName() string { return "goose_db_version" }
