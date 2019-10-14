package dbclickhouse // import "gopkg.in/webnice/kit.v1/modules/dbclickhouse"

//import "gopkg.in/webnice/debug.v1"
//import "gopkg.in/webnice/log.v2"
import (
	"time"

	"gopkg.in/webnice/kit.v1/modules/dbclickhouse/connector"

	"github.com/jmoiron/sqlx"
)

var defaultConfiguration *Configuration

// Interface is an interface of repository
type Interface interface {
	// Dsn Return DSN string
	Dsn() (string, error)

	// Connect Установка соединения с базой данных
	Connect() error

	// Disconnect Закрытие соединения с базой данных
	Disconnect() error

	// Gist Return DB connection object
	Gist() *sqlx.DB

	// Debug Включение или отключение режима отладки
	Debug(d bool)
}

// Implementation Is an implementation of module
type Implementation struct {
	cnf   *Configuration      // Конфигурация подключения к базе данных
	dsn   string              // Строка подключения к базе данных
	debug bool                // Режим дебага
	conn  connector.Interface // Интерфейс соединения с базой данных считающий количество открытий и закрытий
}

// Configuration SQL database configuration structure
type Configuration struct {
	Hosts        []string      `yaml:"Hosts"           json:"hosts"`         // Хосты базы данных
	Login        string        `yaml:"Login"           json:"login"`         // Логин к базе данных
	Password     string        `yaml:"Password"        json:"password"`      // Пароль к базе данных
	Database     string        `yaml:"Database"        json:"database"`      // База данных по умолчанию
	ReadTimeout  time.Duration `yaml:"ReadTimeout"     json:"read_timeout"`  // Таймаут чтения данных из БД
	WriteTimeout time.Duration `yaml:"WriteTimeout"    json:"wtite_timeout"` // Таймаун записи данных в БД
	NoDelay      bool          `yaml:"NoDelay"         json:"no_delay"`      // Флаг для unix socket
	Compress     bool          `yaml:"Compress"        json:"compress"`      // Сжатие данных
	OpenStrategy string        `yaml:"OpenStrategy"    json:"open_strategy"` // Стратегия подключения к хостам. Значения: random, in_order
	BlockSize    uint64        `ysml:"BlockSize"       json:"block_size"`    // Максимальное количество строк в блоке
	Secure       bool          `yaml:"Secure"          json:"secure"`        // Установить безопасное соединение
	SkipVerify   bool          `yaml:"SkipVerify"      json:"skip_verify"`   // Пропустить проверку сертификата
	Debug        bool          `yaml:"Debug"           json:"debug"`         // Режим отладки
	Migrations   string        `yaml:"Migrations"      json:"migrations"`    // Путь к папке с файлами миграций базы данных
}
