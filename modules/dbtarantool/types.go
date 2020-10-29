package dbtarantool // import "github.com/webnice/kit/v1/modules/dbtarantool"

import (
	"time"

	"github.com/webnice/kit/v1/modules/dbtarantool/connector"
	"github.com/webnice/kit/v1/modules/dbtarantool/tarantool"
)

var defaultConfiguration *Configuration

// Interface is an interface of repository
type Interface interface {
	// Connect Установка соединения с базой данных
	Connect() error

	// Disconnect Закрытие соединения с базой данных
	Disconnect() error

	// Gist Return DB connection object
	Gist() *tarantool.Connection

	// Debug Включение или отключение режима отладки
	Debug(d bool)
}

// Implementation Is an implementation of module
type Implementation struct {
	cnf   *Configuration      // Конфигурация подключения к базе данных
	opt   *tarantool.Options  // Настройки подключения к базе данных
	debug bool                // Режим дебага
	conn  connector.Interface // Интерфейс соединения с базой данных считающий количество открытий и закрытий
}

// Configuration Tarantool database configuration structure
type Configuration struct {
	Host           string        `yaml:"Host"           json:"host"`           // An ip address or host name of the database
	Port           uint16        `yaml:"Port"           json:"port"`           // Port connection mode tcp/ip
	Type           string        `yaml:"Type"           json:"type"`           // The type and mode of connection to the database. Possible values: socket, tcp
	Socket         string        `yaml:"Socket"         json:"socket"`         // The path and name of the socket database
	Login          string        `yaml:"Login"          json:"login"`          // Login to connect to the database
	Password       string        `yaml:"Password"       json:"password"`       // Password database connection
	ConnectTimeout time.Duration `yaml:"ConnectTimeout" json:"connectTimeout"` // Waiting time for communication setup
	QueryTimeout   time.Duration `yaml:"QueryTimeout"   json:"queryTimeout"`   // Query timeout
	DefaultSpace   string        `yaml:"DefaultSpace"   json:"defaultSpace"`   // Default database space
	Shared         string        `yaml:"Shared"         json:"shared"`         // Folder with full database access for importing and exporting data (path from tarantool)
	SharedPath     string        `yaml:"SharedPath"     json:"sharedPath"`     // The same 'Shared' folder (path from service)
}
