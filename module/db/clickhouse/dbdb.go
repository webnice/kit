package dbclickhouse

import (
	"github.com/webnice/kit/v2/module/db/clickhouse/connector"
	log "github.com/webnice/lv2"

	"github.com/jmoiron/sqlx"
)

// New creates new lib implementation
func New(cnf ...*Configuration) Interface {
	var db = new(Implementation)

	// Установка явно переданной конфигурации подключения к базе данных
	for i := range cnf {
		if cnf[i] != nil {
			db.cnf = cnf[i]
		}
	}
	// Установка конфигурации подключения к базе данных по умолчанию
	if db.cnf == nil {
		db.cnf = defaultConfiguration
	}
	if db.cnf.Debug {
		db.debug = true
	}

	return db
}

// Default Установка конфигурации подключения к базе данных по умолчанию
func Default(cnf *Configuration) { defaultConfiguration = cnf }

// Debug Включение или отключение режима debug
func (db *Implementation) Debug(d bool) { db.debug = d }

// findConnectorObject Если есть контекст, и в контексте есть сессия работы с базой данных
func (db *Implementation) findConnectorObject() {
	if db.conn == nil {
		db.conn = connector.New()
	}
}

// Connect Установка соединения с базой данных
func (db *Implementation) Connect() (err error) {
	db.findConnectorObject()

	// Already open
	if db.conn.IsOpened() {
		return
	}

	// Подготавливаем новое соединение
	if err = db.MakeDsn(); err != nil {
		log.Errorf("Unable to create a line of DSN: %s", err.Error())
		return
	}

	// Открываем соединение
	err = db.conn.Open(db.dsn)
	if err != nil {
		log.Errorf("Unable to open clickhouse database session Open(%q): %s", db.dsn, err.Error())
		return
	}
	if db.debug {
		log.Debug(" - Database connection openned")
	}

	return
}

// Disconnect Закрытие соединения с базой данных
func (db *Implementation) Disconnect() (err error) {
	db.findConnectorObject()
	if db.conn == nil {
		return
	}
	// Закрываем соединение
	err = db.conn.Close()
	if db.debug {
		log.Debug(" - Database connection closed")
	}
	db.conn = nil

	return
}

// Gist Return DB connection object
func (db *Implementation) Gist() *sqlx.DB {
	db.findConnectorObject()
	// Autoconnect
	if !db.conn.IsOpened() {
		if err := db.Connect(); err != nil {
			log.Criticalf("Unable connect to database from Gist(): %s", err.Error())
			return nil
		}
	}
	return db.conn.Gist()
}
