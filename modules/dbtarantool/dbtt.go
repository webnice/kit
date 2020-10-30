package dbtarantool

import (
	"github.com/webnice/kit/modules/dbtarantool/connector"
	"github.com/webnice/kit/modules/dbtarantool/tarantool"
	"github.com/webnice/log/v2"
)

// New creates new lib implementation
func New(cnf ...*Configuration) Interface {
	var tt *Implementation
	var i int

	tt = new(Implementation)
	// Установка яавно переданной конфигурации подключения к базе данных
	for i = range cnf {
		if cnf[i] != nil {
			tt.cnf = cnf[i]
		}
	}
	// Установка конфигурации подключения к базе данных по умолчанию
	if tt.cnf == nil {
		tt.cnf = defaultConfiguration
	}
	return tt
}

// Default Установка конфигурации подключения к базе данных по умолчанию
func Default(cnf *Configuration) {
	defaultConfiguration = cnf
}

// Debug Включение или отключение режима debug
func (tt *Implementation) Debug(d bool) { tt.debug = d }

// findConnectorObject Если есть контекст, и в контексте есть сессия работы с базой данных
func (tt *Implementation) findConnectorObject() {
	if tt.conn == nil {
		tt.conn = connector.New()
	}
}

// Connect Установка соединения с базой данных
func (tt *Implementation) Connect() (err error) {
	tt.findConnectorObject()

	// Already open
	if tt.conn.IsOpened() {
		return
	}

	// Открываем соединение
	err = tt.conn.Open(tt.MakeConnectArgs())
	if err != nil {
		log.Errorf("Unable to open database session Open(%s, %s, %d, %v): %s", tt.cnf.Type, tt.cnf.Host, tt.cnf.Port, tt.opt, err)
		return
	}
	if tt.debug {
		log.Debug(" - Database connection openned")
	}

	return
}

// Disconnect Закрытие соединения с базой данных
func (tt *Implementation) Disconnect() (err error) {
	tt.findConnectorObject()
	if tt.conn == nil {
		return
	}
	// Закрываем соединение
	err = tt.conn.Close()
	if tt.debug {
		log.Debug(" - Database connection closed")
	}
	tt.conn = nil

	return
}

// Gist Return ORM object
func (tt *Implementation) Gist() *tarantool.Connection {
	tt.findConnectorObject()
	// Autoconnect
	if !tt.conn.IsOpened() {
		if err := tt.Connect(); err != nil {
			log.Criticalf("Unable connect to database from Gist(): %s", err.Error())
			return nil
		}
	}
	return tt.conn.Gist()
}
