package db

import (
	"github.com/webnice/kit/modules/db/connector"

	"gorm.io/gorm"
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
	Gist() *gorm.DB

	// Debug Включение или отключение режима отладки
	Debug(d bool)
}

// Implementation Is an implementation of module
type Implementation struct {
	cnf   *Configuration      // Конфигурация подключения к базе данных
	drv   string              // Драйвер базы данных
	dsn   string              // Строка подключения к базе данных
	debug bool                // Режим дебага
	conn  connector.Interface // Интерфейс соединения с базой данных считающий количество открытий и закрытий
}

// Configuration SQL database configuration structure
type Configuration struct {
	Driver     string `yaml:"Driver"      json:"driver"`     // Драйвер
	Host       string `yaml:"Host"        json:"host"`       // Хост базы данных
	Port       int16  `yaml:"Port"        json:"port"`       // Порт подключения по протоколу tcp/ip
	Type       string `yaml:"Type"        json:"type"`       // Тип подключения к базе данных socket | tcp
	Socket     string `yaml:"Socket"      json:"socket"`     // Путь к socket файлу
	Name       string `yaml:"Name"        json:"name"`       // Имя базы данных
	Login      string `yaml:"Login"       json:"login"`      // Логин к базе данных
	Password   string `yaml:"Password"    json:"password"`   // Пароль к базе данных
	Charset    string `yaml:"Charset"     json:"charset"`    // Кодировка данных
	Migrations string `yaml:"Migrations"  json:"migrations"` // Путь к папке с файлами миграций базы данных
}
