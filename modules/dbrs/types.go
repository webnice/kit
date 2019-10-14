package dbrs // import "gopkg.in/webnice/kit.v1/modules/dbrs"

//import "gopkg.in/webnice/debug.v1"
//import "gopkg.in/webnice/log.v2"
import (
	"time"

	"gopkg.in/webnice/kit.v1/modules/dbrs/connector"

	redis "github.com/go-redis/redis/v7"
)

var defaultConfiguration *Configuration

// Interface is an interface of repository
type Interface interface {
	// Connect Установка соединения с базой данных
	Connect() error

	// Disconnect Закрытие соединения с базой данных
	Disconnect() error

	// Gist Return DB connection object
	Gist() *redis.Client

	// Debug Включение или отключение режима отладки
	Debug(d bool)
}

// Implementation Is an implementation of module
type Implementation struct {
	cnf   *Configuration      // Конфигурация подключения к базе данных
	opt   *redis.Options      // Настройки подключения к базе данных
	debug bool                // Режим дебага
	conn  connector.Interface // Интерфейс соединения с базой данных считающий количество открытий и закрытий
}

// Configuration Redis database configuration structure
type Configuration struct {
	Host               string        `yaml:"Host"               json:"host"`               // An ip address or host name of the database
	Port               uint16        `yaml:"Port"               json:"port"`               // Port connection mode tcp/ip. Default 6379
	Type               string        `yaml:"Type"               json:"type"`               // The type and mode of connection to the database. Possible values: socket, tcp. Default is tcp
	Socket             string        `yaml:"Socket"             json:"socket"`             // The path and name of the database socket
	Password           string        `yaml:"Password"           json:"password"`           // Password database connection
	Database           int64         `yaml:"Database"           json:"database"`           // Database number, =0 for default database
	MaxRetries         int64         `yaml:"MaxRetries"         json:"maxRetries"`         // Maximum number of retries before giving up
	MinRetryBackoff    time.Duration `yaml:"MinRetryBackoff"    json:"minRetryBackoff"`    // Minimum backoff between each retry. Default is 8 milliseconds. =-1 disables backoff
	MaxRetryBackoff    time.Duration `jaml:"MaxRetryBackoff"    json:"maxRetryBackoff"`    // Maximum backoff between each retry. Default is 512 milliseconds. =-1 disables backoff
	DialTimeout        time.Duration `jaml:"DialTimeout"        json:"dialTimeout"`        // Dial timeout for establishing new connections. Default is 5 seconds
	ReadTimeout        time.Duration `jaml:"ReadTimeout"        json:"readTimeout"`        // Timeout for socket reads. If reached, commands will fail with a timeout instead of blocking. Use value -1 for no timeout and 0 for default. Default is 3 seconds
	WriteTimeout       time.Duration `jaml:"WriteTimeout"       json:"writeTimeout"`       // Timeout for socket writes. If reached, commands will fail with a timeout instead of blocking. Default is ReadTimeout
	PoolSize           int64         `jaml:"PoolSize"           json:"poolSize"`           // Maximum number of socket connections. Default is 10 connections per every CPU as reported by runtime.NumCPU
	MinIdleConns       int64         `jaml:"MinIdleConns"       json:"minIdleConns"`       // Minimum number of idle connections which is useful when establishing new connection is slow
	MaxConnAge         time.Duration `jaml:"MaxConnAge"         json:"maxConnAge"`         // Connection age at which client retires (closes) the connection. Default is to not close aged connections
	PoolTimeout        time.Duration `jaml:"PoolTimeout"        json:"poolTimeout"`        // Amount of time client waits for connection if all connections are busy before returning an error. Default is ReadTimeout + 1 second
	IdleTimeout        time.Duration `jaml:"IdleTimeout"        json:"idleTimeout"`        // Amount of time after which client closes idle connections. Should be less than server's timeout. Default is 5 minutes. -1 disables idle timeout check
	IdleCheckFrequency time.Duration `jaml:"IdleCheckFrequency" json:"idleCheckFrequency"` // Frequency of idle checks made by idle connections reaper. Default is 1 minute. -1 disables idle connections reaper, but idle connections are still discarded by the client if IdleTimeout is set
}
