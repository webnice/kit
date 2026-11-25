package gelf

import (
	"crypto/tls"
	"io"
	"net"
	"sync"
	"time"
)

// Разделитель сообщений протоколов TCP и TLS.
var messageSeparator = []byte{0}

// Compression Тип сжатия данных сообщений.
type Compression uint8

// Настройки клиента отправки сообщений GELF.
type options struct {
	compressLevel  int           // Уровень сжатия сообщений, константы пакета compress/flate.
	chunkSize      uint          // Размер полезных данных одного чанка UDP сообщения.
	dialTimeout    time.Duration // Тайм-аут установки сетевого соединения, ноль - без ограничения.
	maxReconnect   int           // Максимальное количество попыток переподключения при обрыве связи.
	reconnectDelay time.Duration // Пауза между попытками переподключения.
	tlsConfig      *tls.Config   // Конфигурация TLS для протокола tls.
}

// Option Функция настройки клиента отправки сообщений GELF.
type Option func(opt *options)

// ProtocolClient Интерфейс клиента сетевого протокола отправки данных сообщений.
type ProtocolClient interface {
	// SendData Отправка подготовленных данных сообщения серверу.
	SendData(data []byte) (err error)

	// Close Закрытие соединения и освобождение занятых ресурсов.
	Close() (err error)
}

// Client Интерфейс клиента отправки сообщений GELF.
type Client interface {
	ProtocolClient
	io.Writer

	// Send Кодирование и отправка сообщения серверу GELF.
	Send(msg *Message) (err error)
}

// Message Сообщение формата GELF версии 1.1.
type Message struct {
	Version  string  `json:"version"`                // Версия формата сообщения.
	Host     string  `json:"host"`                   // Имя хоста отправителя сообщения.
	Short    string  `json:"short_message"`          // Краткий текст сообщения.
	Full     string  `json:"full_message,omitempty"` // Полный текст сообщения.
	Time     float64 `json:"timestamp"`              // Временная метка UNIX.
	Level    int32   `json:"level"`                  // Уровень важности сообщения, syslog severity.
	Facility string  `json:"facility,omitempty"`     // Категория (подсистема) отправителя сообщения.

	// Fields Дополнительные поля сообщения верхнего уровня. При кодировании имена полей передаются без изменений.
	Fields map[string]any `json:"-"`

	// Keys Дополнительные поля сообщения. При кодировании имена полей получают префикс "_".
	Keys map[string]any `json:"-"`
}

// Клиент отправки сообщений GELF.
type client struct {
	protocol      ProtocolClient // Клиент сетевого протокола отправки данных.
	compression   Compression    // Тип сжатия данных сообщений.
	compressLevel int            // Уровень сжатия сообщений.
	hostname      string         // Имя хоста отправителя сообщений.
	facility      string         // Категория (подсистема) отправителя сообщений.
}

// Клиент протокола UDP.
type udpClient struct {
	conn      *net.UDPConn // Постоянное соединение с сервером получателем сообщений.
	chunkSize uint         // Размер полезных данных одного чанка.
}

// Клиент протокола TCP.
type tcpClient struct {
	conn *tcpConnMng // Менеджер соединения с сервером получателем сообщений.
}

// Менеджер соединения с сервером с автоматическим переподключением при обрыве связи.
type tcpConnMng struct {
	sync.RWMutex                            // Защита активного соединения.
	wmu            sync.Mutex               // Сериализация записи сообщений.
	dial           func() (net.Conn, error) // Функция установки соединения с сервером.
	maxReconnect   int                      // Максимальное количество попыток переподключения при обрыве связи.
	reconnectDelay time.Duration            // Пауза между попытками переподключения.
	conn           net.Conn                 // Активное соединение с сервером.
}
