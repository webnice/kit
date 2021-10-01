package dbpostgres

//import "github.com/webnice/kit/modules/dbpostgres/connector"

//var defaultConfiguration *Configuration
//
//// Interface is an interface of repository
//type Interface interface {
//	// Connect Установка соединения с базой данных
//	Connect() error
//
//	// Disconnect Закрытие соединения с базой данных
//	Disconnect() error
//
//	// Gist Return DB connection object
//	Gist() *tarantool.Connection
//
//	// Debug Включение или отключение режима отладки
//	Debug(d bool)
//}
//
//// Implementation Is an implementation of module
//type Implementation struct {
//	cnf   *Configuration      // Конфигурация подключения к базе данных
//	opt   *tarantool.Options  // Настройки подключения к базе данных
//	debug bool                // Режим дебага
//	conn  connector.Interface // Интерфейс соединения с базой данных считающий количество открытий и закрытий
//}

// Configuration PostgreSQL database configuration structure
type Configuration struct {
	Driver     string `yaml:"Driver"      json:"driver"`     // Драйвер базы данных
	Host       string `yaml:"Host"        json:"host"`       // Хост базы данных
	Port       uint16 `yaml:"Port"        json:"port"`       // Порт подключения по протоколу tcp/ip
	Type       string `yaml:"Type"        json:"type"`       // Тип подключения к базе данных socket | tcp
	Socket     string `yaml:"Socket"      json:"socket"`     // Путь к socket файлу
	Name       string `yaml:"Name"        json:"name"`       // Название базы данных
	Login      string `yaml:"Login"       json:"login"`      // Логин к базе данных
	Password   string `yaml:"Password"    json:"password"`   // Пароль к базе данных
	Charset    string `yaml:"Charset"     json:"charset"`    // Кодировка данных
	Migrations string `yaml:"Migrations"  json:"migrations"` // Путь к папке с файлами миграций базы данных
}

/*

pgsql:host=localhost;port=5432;dbname=testdb;user=bruce;password=mypass

postgresql://[user[:password]@][netloc][:port][/dbname][?param1=value1&...]`
pgsql://username@%Fvar%Frun%Fpostgresql%F.s.PGSQL.6432/dbname
postgresql://%2Fvar%2Flib%2Fpostgresql/dbname
postgres://username@/dbname
postgresql://user@host/database?socket=/path/to/socket
postgres://user:password@/database?host=/path/to/socket/dir
postgres://user:password@/database

postgres://mmuser:mmuser_password@127.0.0.1:5432/mattermostdb?sslmode=disable&connect_timeout=10
postgres:///mattermostdb?host=/run/postgresql


*/
