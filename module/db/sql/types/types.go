package types

import (
	"time"

	kmll "github.com/webnice/kit/v4/module/log/level"
)

// Configuration SQL database configuration structure.
type Configuration struct {
	// Driver Драйвер.
	Driver string `yaml:"Driver"                                                    default-value:"mysql"`
	// Host Хост базы данных.
	Host string `yaml:"Host"                                                        default-value:"localhost"`
	// Port Порт подключения по протоколу tcp/ip.
	Port uint16 `yaml:"Port"                                                        default-value:"3306"`
	// Type Тип подключения к базе данных socket | tcp.
	Type string `yaml:"Type"                                                        default-value:"tcp"`
	// Socket Путь к socket файлу.
	Socket string `yaml:"Socket"                                                    default-value:"-"`
	// Name Имя базы данных.
	Name string `yaml:"Name"                                                        default-value:"database"`
	// Login Логин к базе данных.
	Login string `yaml:"Login"                                                      default-value:"root"`
	// Password Пароль к базе данных.
	Password string `yaml:"Password"                                                default-value:"-"`
	// Migration Путь к папке с файлами миграций базы данных.
	Migration string `yaml:"Migration"                                              default-value:"-"`

	// НАСТРОЙКИ СОЕДИНЕНИЯ И БИБЛИОТЕК

	// Charset Кодировка данных.
	Charset string `yaml:"Charset"                                                  default-value:"utf8"`
	// ParseTime Парсинг значений даты и времени.
	ParseTime bool `yaml:"ParseTime"                                                default-value:"true"`
	// TimezoneLocation Зона времени по умолчанию.
	TimezoneLocation string `yaml:"TimezoneLocation"                                default-value:"Local"`
	// DefaultStringSize Размер значений по умолчанию для строковых колонок.
	DefaultStringSize uint `yaml:"DefaultStringSize"                                default-value:"256"`
	// CreateBatchSize Размер пакета групповой вставки по умолчанию.
	CreateBatchSize int `yaml:"CreateBatchSize"                                     default-value:"100"`
	// DisableDatetimePrecision Отключить точность datetime колонок для совместимости с MySQL 5.6 и более старой.
	DisableDatetimePrecision bool `yaml:"DisableDatetimePrecision"                  default-value:"false"`
	// MaxIdleConn Максимальное количество соединений в пуле бездействия.
	MaxIdleConn int `yaml:"MaxIdleConn"                                             default-value:"10"`
	// MaxOpenConn Максимальное количество открытых соединений с БД.
	MaxOpenConn int `yaml:"MaxOpenConn"                                             default-value:"20"`
	// MaxIdleTimeConn Время ожидания не используемого соединения перед закрытием.
	MaxIdleTimeConn time.Duration `yaml:"MaxIdleTimeConn"                           default-value:"5m"`
	// MaxLifetimeConn Максимальное время повторного использования соединения.
	MaxLifetimeConn time.Duration `yaml:"MaxLifetimeConn"                           default-value:"1h"`
	// SkipDefaultTransaction Не создавать транзакции для запросов к базе данных.
	SkipDefaultTransaction bool `yaml:"SkipDefaultTransaction"                      default-value:"false"`
	// DisableAutomaticPing Отключает автоматический пинг перед запросом к базе данных.
	DisableAutomaticPing bool `yaml:"DisableAutomaticPing"                          default-value:"false"`
	// PrepareStmt Включается подготовка данных и кеширование их при выполнении любого SQL запроса.
	PrepareStmt bool `yaml:"PrepareStmt"                                            default-value:"false"`
	// PostgreSQLPreferSimpleProtocol отключает неявное использование подготовленных инструкций.
	// По умолчанию pgx автоматически использует расширенный протокол. Это может повысить производительность за
	// счёт возможности использования двоичного формата. Он также не зависит от очистки параметров на стороне клиента.
	// Так же, он требует двух обходов для каждого запроса (если не используется подготовленный оператор) и может быть
	// несовместимым с прокси-серверами, такими как PGBouncer.
	// Установка параметра PostgreSQLPreferSimpleProtocol=true приводит к тому, что по умолчанию используется простой
	// протокол.
	PostgreSQLPreferSimpleProtocol bool `yaml:"PostgreSQLPreferSimpleProtocol"      default-value:"true"`

	// ЖУРНАЛИРОВАНИЕ ЗАПРОСОВ

	// Loglevel Уровень логирования SQL запросов.
	// Драйвером базы данных и ОРМ используется ограниченное количество уровней логирования, все остальные уровни
	// логирования игнорируются. Активные уровни перечислены ниже:
	// error   - Запросы, выполнение которых завершилось ошибкой.
	// warning - Запросы с ошибками, а так же требующие повышенного внимания, но не являющиеся ошибкой.
	// info    - Все без исключения запросы к базе данных.
	// Значением по умолчанию является off - всё логирование отключено.
	Loglevel kmll.Level `yaml:"Loglevel"                                            default-value:"off"`
}
