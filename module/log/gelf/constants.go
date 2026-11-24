package gelf

import (
	"crypto/tls"
	"time"
)

// DefaultChunkSize Размер полезных данных одного чанка UDP сообщения по умолчанию.
// Значение подобрано с учётом типичного размера MTU сетевого сегмента.
const DefaultChunkSize uint = 1420

// Значения по умолчанию политики переподключения протоколов TCP и TLS.
const (
	// DefaultMaxReconnect Максимальное количество попыток переподключения к серверу
	// при обрыве связи по умолчанию.
	DefaultMaxReconnect = 3

	// DefaultReconnectDelay Пауза между попытками переподключения к серверу по умолчанию.
	DefaultReconnectDelay = time.Second
)

// Уровни важности сообщения GELF, соответствуют syslog severity.
const (
	// SeverityEmergency Система неработоспособна.
	SeverityEmergency int32 = iota

	// SeverityAlert Действие должно быть выполнено немедленно.
	SeverityAlert

	// SeverityCritical Критическое состояние системы.
	SeverityCritical

	// SeverityError Ошибочное состояние.
	SeverityError

	// SeverityWarning Предупреждение.
	SeverityWarning

	// SeverityNotice Нормальное, но значимое состояние.
	SeverityNotice

	// SeverityInfo Информационное сообщение.
	SeverityInfo

	// SeverityDebug Отладочное сообщение.
	SeverityDebug
)

// Типы сжатия данных сообщений.
const (
	// CompressionNone Отправка сообщений без сжатия.
	CompressionNone Compression = iota

	// CompressionGzip Сжатие сообщений алгоритмом gzip.
	CompressionGzip

	// CompressionZlib Сжатие сообщений алгоритмом zlib.
	CompressionZlib
)

// Версия формата сообщений GELF.
const version = "1.1"

// Префикс дополнительных полей сообщения GELF.
const keyPrefix = "_"

// Поддерживаемые сетевые протоколы.
const (
	protoUdp    = "udp"
	protoTcp    = "tcp"
	protoTls    = "tls"
	protoTcpTls = "tcp+tls"
)

// Названия типов сжатия сообщений.
const (
	compressionNameNone = ""
	compressionNameGzip = "gzip"
	compressionNameZlib = "zlib"
)

// Константы протокола UDP.
const (
	udpNetwork      = "udp"                                         // Название сетевого протокола.
	chunkMagicSize  = 2                                             // Размер магического заголовка чанка.
	chunkMagic0     = 0x1e                                          // Первый байт магического заголовка чанка.
	chunkMagic1     = 0x0f                                          // Второй байт магического заголовка чанка.
	messageIDSize   = 8                                             // Размер идентификатора сообщения в чанке.
	sequenceSize    = 2                                             // Размер блока последовательности чанка.
	chunkHeaderSize = chunkMagicSize + messageIDSize + sequenceSize // Размер заголовка чанка.
	maxChunkCount   = 128                                           // Максимальное количество чанков в одном сообщении.
)

// Константы протоколов TCP и TLS.
const (
	tcpNetwork    = "tcp"            // Название сетевого протокола TCP.
	tlsMinVersion = tls.VersionTLS12 // Минимальная версия TLS по умолчанию.
)
