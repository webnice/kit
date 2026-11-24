package gelf

import (
	"compress/flate"
	"crypto/tls"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// New Создание клиента GELF по адресу сервера.
// Формат адреса: [proto://]host:port, где proto - udp (значение по умолчанию), tcp, tls или tcp+tls.
func New(address string, compression Compression, opts ...Option) (ret Client, err error) {
	var (
		opt   options
		apply Option
		proto string
		host  string
		port  uint16
		cli   *client
	)

	// Проверка типа сжатия сообщений.
	switch compression {
	case CompressionNone, CompressionGzip, CompressionZlib:
	default:
		err = Errors().UnknownCompression.Bind(compression)
		return
	}
	// Разбор адреса сервера.
	if proto, host, port, err = parseAddress(address); err != nil {
		return
	}
	// Проверка поддержки сжатия сообщений выбранным протоколом.
	if proto != protoUdp && compression != CompressionNone {
		err = Errors().CompressionNotSupported.Bind(compressionTitle(compression), proto)
		return
	}
	// Применение настроек клиента.
	opt = defaultOptions()
	for _, apply = range opts {
		apply(&opt)
	}
	// Создание клиента отправки сообщений.
	cli = &client{
		compression:   compression,
		compressLevel: opt.compressLevel,
		facility:      filepath.Base(os.Args[0]),
	}
	cli.hostname, _ = os.Hostname()
	// Создание клиента сетевого протокола.
	switch proto {
	case protoTcp:
		if cli.protocol, err = newTcpClient(host, port, opt, false); err != nil {
			return
		}
	case protoTls:
		if cli.protocol, err = newTcpClient(host, port, opt, true); err != nil {
			return
		}
	default:
		if cli.protocol, err = newUdpClient(host, port, opt); err != nil {
			return
		}
	}
	ret = cli

	return
}

// WithCompressLevel Уровень сжатия сообщений gzip или zlib, константы пакета compress/flate.
func WithCompressLevel(level int) (ret Option) {
	ret = func(opt *options) { opt.compressLevel = level }

	return
}

// WithChunkSize Размер полезных данных одного чанка UDP сообщения.
func WithChunkSize(size uint) (ret Option) {
	ret = func(opt *options) { opt.chunkSize = size }

	return
}

// WithDialTimeout Тайм-аут установки сетевого соединения с сервером.
func WithDialTimeout(timeout time.Duration) (ret Option) {
	ret = func(opt *options) { opt.dialTimeout = timeout }

	return
}

// WithReconnect Политика переподключения протоколов TCP и TLS при обрыве связи:
// максимальное количество попыток и пауза между ними.
func WithReconnect(attempts int, delay time.Duration) (ret Option) {
	ret = func(opt *options) { opt.maxReconnect, opt.reconnectDelay = attempts, delay }

	return
}

// WithTLSConfig Конфигурация TLS для протокола tls. Если не указана, создаётся
// конфигурация по умолчанию с именем сервера из адреса и минимальной версией TLS 1.2.
func WithTLSConfig(cfg *tls.Config) (ret Option) {
	ret = func(opt *options) { opt.tlsConfig = cfg }

	return
}

// Настройки клиента по умолчанию.
func defaultOptions() (ret options) {
	ret = options{
		compressLevel:  flate.DefaultCompression,
		chunkSize:      DefaultChunkSize,
		maxReconnect:   DefaultMaxReconnect,
		reconnectDelay: DefaultReconnectDelay,
	}

	return
}

// Название типа сжатия сообщений.
func compressionTitle(compression Compression) (ret string) {
	switch compression {
	case CompressionGzip:
		ret = compressionNameGzip
	case CompressionZlib:
		ret = compressionNameZlib
	default:
		ret = compressionNameNone
	}

	return
}

// Разбор адреса сервера в формате [proto://]host:port.
func parseAddress(address string) (proto, host string, port uint16, err error) {
	const (
		errTextNoPort        = "не указан порт подключения"
		errTextPortNotNumber = "порт подключения не является числом"
	)
	var (
		uri *url.URL
		prt uint64
	)

	// Дополнение адреса протоколом по умолчанию.
	if !strings.Contains(address, "://") {
		address = protoUdp + "://" + address
	}
	// Разбор адреса и проверка протокола.
	if uri, err = url.Parse(address); err != nil {
		err = Errors().InvalidAddress.Bind(address, err.Error())
		return
	}
	switch strings.ToLower(uri.Scheme) {
	case protoUdp, protoTcp:
		proto = strings.ToLower(uri.Scheme)
	case protoTls, protoTcpTls:
		proto = protoTls
	default:
		err = Errors().UnknownProtocol.Bind(uri.Scheme)
		return
	}
	// Проверка порта подключения.
	host = uri.Hostname()
	if uri.Port() == "" {
		err = Errors().InvalidAddress.Bind(address, errTextNoPort)
		return
	}
	if prt, err = strconv.ParseUint(uri.Port(), 10, 16); err != nil {
		err = Errors().InvalidAddress.Bind(address, errTextPortNotNumber)
		return
	}
	port = uint16(prt)

	return
}
