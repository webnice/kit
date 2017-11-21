package db // import "gopkg.in/webnice/kit.v1/modules/db"

//import "gopkg.in/webnice/debug.v1"
import "gopkg.in/webnice/log.v2"
import (
	"gopkg.in/webnice/kit.v1/modules/db/connector"

	"github.com/jinzhu/gorm"

	// gorm dependences
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// New creates new repository implementation
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

	// Устанавливаем свой собственный логер
	//db.Logger()

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

	//log.Debug(" - Database connection closed")
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
	// Debug
	//	if configuration.Get().Debug() {
	//		return db.conn.Gist().Debug()
	//	}
	return db.conn.Gist()
}
