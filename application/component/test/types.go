// Package test
package test

import (
	"context"
	"net"
	"time"

	kitModuleCfg "github.com/webnice/kit/module/cfg"
	kmll "github.com/webnice/kit/module/log/level"
)

// Структура объекта компоненты.
type impl struct {
	ctx    context.Context
	cfn    context.CancelFunc
	Dumper func(...interface{})
	cfg    kitModuleCfg.Interface
	opt    *options
	cobj   *Configuration
}

type options struct {
	ArgOne string `arg:"" help:"Аргумент первый."`
	Fast   bool   `cmd:"" help:"Быстро." short:"f"`
	Slow   string `cmd:"" help:"Медленно." short:"s"`
	Flag   string `cmd:"" help:"Какой-то флаг." short:"g"`
}

// Configuration Конфигурация компоненты
type Configuration struct {
	NotHomeDirectory string         `yaml:"NotHomeDirectory"  env-name:"NOT_HOME_DIRECTORY"  default-value:"~"                                                description:"Моя тестовая домашняя директория."`
	MyLogLevel       kmll.Level     `yaml:"MyLogLevel"        env-name:"MY_LOG_LEVEL"        default-value:"info"                                             description:"Тестовый\nМногострочный\nКомментарий."`
	TestLocation     *time.Location `yaml:"TestLocation"                                     default-value:"Europe/Moscow"`
	TestExotic       string         `yaml:"TestExotic"`
	IP               net.IP         `yaml:"IP"                                               default-value:"127.0.0.1"                                        description:"IPv4 адрес."`
	PostgreSql       Psql           `yaml:"psql"                                                                                                              description:"Реквизиты подключения к базе данных PostgreSQL."`
	//Database         Database       `yaml:"Database"                                                                                                          description:"Реквизиты подключения к базе данных."`
	//TestField        time.Time         `yaml:"TestField"         env-name:"-"                   default-value:"-"                                                description:"-"                                      `
	//TestValue        time.Time         `yaml:"Test Value"        env-name:"-"                   default-value:"2022-05-11T07:40:00Z"                             description:"Просто тестовая переменная."`
	//TestDuration     time.Duration     `yaml:"TestDuration"                                     default-value:"1h9s"`
	//TestDurationAdr  *time.Duration    `yaml:"TestDurationAdr"                                  default-value:"1h9s"`
	//SliceBool        []bool            `yaml:"slice_bool"        env-name:"-"                   default-value:"true,false,true,true"                             description:""`
	//SliceInt         []*int            `yaml:"slice_int"         env-name:"-"                   default-value:"4,0,3,2,1"                                        description:""`
	//SliceString      []string          `yaml:"slice_string"      env-name:"-"                   default-value:"111 222 333 444 000 555"                          description:""`
	//SliceInt64       [5]int64          `yaml:"slice_int64"       env-name:"-"                   default-value:"4,0,321,4,5,6"                                    description:""`
	//MapInt64         map[string]uint16 `yaml:"mapInt64"          env-name:"-"                   default-value:"первое=1,второе значение=2, третье значение = 3"  description:""`
	//MapI             map[uint]string   `yaml:"mapI"              env-name:"-"                   default-value:"0=Ноль,1=Один, 2 = Два, 3=       Тридцать три"    description:""`
}

// Database Конфигурация подключения к базе данных
type Database struct {
	Driver     string     `yaml:"Driver"      json:"driver"                                 default-value:"mysql"`     // Драйвер
	Host       string     `yaml:"Host"        json:"host"        env-name:"MYSQL_HOST"      default-value:"localhost"` // Хост базы данных
	Port       int16      `yaml:"Port"        json:"port"        env-name:"MYSQL_PORT"      default-value:"3306"`      // Порт подключения по протоколу tcp/ip
	Type       string     `yaml:"Type"        json:"type"        env-name:"MYSQL_TYPE"      default-value:"tcp"`       // Тип подключения к базе данных socket | tcp
	Socket     string     `yaml:"Socket"      json:"socket"      env-name:"MYSQL_SOCKET"`                              // Путь к socket файлу
	Name       string     `yaml:"Name"        json:"name"        env-name:""                default-value:"database"`  // Имя базы данных
	Login      string     `yaml:"Login"       json:"login"       env-name:"MYSQL_LOGIN"     default-value:"root"`      // Логин к базе данных
	Password   string     `yaml:"Password"    json:"password"    env-name:"MYSQL_PASSWORD"`                            // Пароль к базе данных
	Charset    string     `yaml:"Charset"     json:"charset"     env-name:""                default-value:"utf8mb4"`   // Кодировка данных
	Migrations string     `yaml:"Migrations"  json:"migrations"  env-name:""`                                          // Путь к папке с файлами миграций базы данных
	Loglevel   kmll.Level `yaml:"Loglevel"    json:"loglevel"    env-name:""                default-value:"debug"`     // Уровень логирования SQL запросов
}

// Psql Конфигурация PostgreSQL
type Psql struct {
	Username           string   `yaml:"user"`
	Password           string   `yaml:"pass"`
	Host               string   `yaml:"host"`
	Port               int16    `yaml:"port"`
	Name               string   `yaml:"dbname"`
	IsSsl              bool     `yaml:"sslmode"`
	MaxIdleConnections uint64   `yaml:"max_idle_conns"`
	MaxOpenConnections uint64   `yaml:"max_open_conns"`
	Migrations         string   `yaml:"migrations"`
	Blacklist          []string `yaml:"blacklist"`
}

// Default Реализация интерфейса types.ConfigurationDefaulter
func (psql *Psql) Default() (err error) {
	const (
		defaultUsername         = "postgres"
		defaultPassword         = "postgres"
		defaultHost             = "localhost"
		defaultPort             = 5432
		defaultDatabaseName     = "database"
		defaultMaxIdle          = 50
		defaultMaxOpen          = 50
		blacklistDbTest         = "test"
		blacklistDbMigrations   = "gopg_migrations"
		blacklistDbGooseVersion = "goose_db_version"
	)
	psql.Username, psql.Password, psql.Host, psql.Port, psql.Name =
		defaultUsername, defaultPassword, defaultHost, defaultPort, defaultDatabaseName
	psql.MaxIdleConnections, psql.MaxOpenConnections = defaultMaxIdle, defaultMaxOpen
	psql.Blacklist = []string{blacklistDbTest, blacklistDbMigrations, blacklistDbGooseVersion}

	return
}
