package dbredis // import "github.com/webnice/kit/v1/modules/dbredis"

import (
	"github.com/webnice/kit/v1/modules/dbredis/connector"
	"github.com/webnice/log/v2"

	redis "github.com/go-redis/redis/v7"
)

// New creates new lib implementation
func New(cnf ...*Configuration) Interface {
	var (
		tt *Implementation
		i  int
	)

	tt = new(Implementation)
	// Установка явно переданной конфигурации подключения к базе данных
	for i = range cnf {
		if cnf[i] != nil {
			tt.cnf = cnf[i]
		}
	}
	// Установка конфигурации подключения к базе данных по умолчанию
	if tt.cnf == nil {
		tt.cnf = defaultConfiguration
	}
	Defaults(tt.cnf)

	return tt
}

// Default Установка конфигурации подключения к базе данных по умолчанию
func Default(cnf *Configuration) { defaultConfiguration = cnf }

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
	if err = tt.conn.Open(tt.MakeConnectArgs()); err != nil {
		log.Errorf("unable to open database session Open(%s, %s, %d, %v): %s", tt.cnf.Type, tt.cnf.Host, tt.cnf.Port, tt.opt, err)
		return
	}
	if tt.debug {
		log.Debug("- database connection opened")
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
		log.Debug("- database connection closed")
	}
	tt.conn = nil

	return
}

// Gist Return ORM object
func (tt *Implementation) Gist() *redis.Client {
	tt.findConnectorObject()
	// Auto connect
	if !tt.conn.IsOpened() {
		if err := tt.Connect(); err != nil {
			log.Criticalf("unable connect to database from Gist(): %s", err)
			return nil
		}
	}
	return tt.conn.Gist()
}
