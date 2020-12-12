package db

import (
	"github.com/webnice/kit/modules/db/connector"
	log "github.com/webnice/lv2"

	"github.com/jinzhu/gorm"

	// gorm dependences
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// New creates new lib implementation
func New(cnf ...*Configuration) Interface {
	var db *Implementation
	var i int

	db = new(Implementation)
	// Установка яавно переданной конфигурации подключения к базе данных
	for i = range cnf {
		if cnf[i] != nil {
			db.cnf = cnf[i]
		}
	}
	// Установка конфигурации подключения к базе данных по умолчанию
	if db.cnf == nil {
		db.cnf = defaultConfiguration
	}
	return db
}

// Default Установка конфигурации подключения к базе данных по умолчанию
func Default(cnf *Configuration) {
	defaultConfiguration = cnf
}

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
	err = db.conn.Open(db.drv, db.dsn)
	if err != nil {
		log.Errorf("Unable to open database session gorm.Open(%s, %s): %s", db.drv, db.dsn, err.Error())
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

// Gist Return ORM object
func (db *Implementation) Gist() *gorm.DB {
	db.findConnectorObject()
	// Autoconnect
	if !db.conn.IsOpened() {
		if err := db.Connect(); err != nil {
			log.Criticalf("Unable connect to database from Gist(): %s", err.Error())
			return nil
		}
	}
	if db.debug {
		return db.conn.Gist().Debug()
	} else {
		return db.conn.Gist()
	}
}
