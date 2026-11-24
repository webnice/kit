package gelf

import (
	"crypto/tls"
	"net"
	"strconv"
	"time"
)

// NewTcp Создание клиента протокола TCP.
func NewTcp(host string, port uint16, opts ...Option) (ret ProtocolClient, err error) {
	var (
		opt   options
		apply Option
	)

	// Применение настроек клиента.
	opt = defaultOptions()
	for _, apply = range opts {
		apply(&opt)
	}
	// Создание клиента протокола TCP.
	if ret, err = newTcpClient(host, port, opt, false); err != nil {
		return
	}

	return
}

// Создание клиента протокола TCP или TLS (useTLS) и установка первого соединения с сервером.
func newTcpClient(host string, port uint16, opt options, useTLS bool) (ret ProtocolClient, err error) {
	var (
		address string
		dialer  *net.Dialer
		conn    *tcpConnMng
	)

	// Создание менеджера соединения с функцией установки соединения.
	address = net.JoinHostPort(host, strconv.FormatUint(uint64(port), 10))
	dialer = &net.Dialer{Timeout: opt.dialTimeout}
	conn = &tcpConnMng{
		maxReconnect:   opt.maxReconnect,
		reconnectDelay: opt.reconnectDelay,
	}
	if useTLS {
		conn.dial = newTlsDialer(address, host, dialer, opt.tlsConfig)
	} else {
		conn.dial = func() (net.Conn, error) { return dialer.Dial(tcpNetwork, address) }
	}
	// Установка первого соединения с сервером.
	if err = conn.reconnect(); err != nil {
		return
	}
	ret = &tcpClient{conn: conn}

	return
}

// Создание функции установки TLS соединения с сервером. Если конфигурация TLS
// не передана, используется конфигурация по умолчанию с именем сервера и минимальной версией TLS 1.2.
func newTlsDialer(address, host string, dialer *net.Dialer, cfg *tls.Config) (ret func() (net.Conn, error)) {
	var tdl *tls.Dialer

	// Конфигурация TLS по умолчанию.
	if cfg == nil {
		cfg = &tls.Config{
			ServerName: host,
			MinVersion: tlsMinVersion,
		}
	}
	// Создание функции установки TLS соединения.
	tdl = &tls.Dialer{
		NetDialer: dialer,
		Config:    cfg,
	}
	ret = func() (net.Conn, error) { return tdl.Dial(tcpNetwork, address) }

	return
}

// SendData Отправка данных сообщения с разделителем сообщений.
func (cli *tcpClient) SendData(data []byte) (err error) {
	var message []byte

	// Формирование сообщения с разделителем.
	message = make([]byte, 0, len(data)+len(messageSeparator))
	message = append(message, data...)
	message = append(message, messageSeparator...)
	// Отправка сообщения серверу.
	if _, err = cli.conn.Write(message); err != nil {
		return
	}

	return
}

// Close Закрытие соединения с сервером.
func (cli *tcpClient) Close() (err error) {
	err = cli.conn.Close()

	return
}

// Получение активного соединения с сервером.
func (tcm *tcpConnMng) connection() (ret net.Conn) {
	tcm.RLock()
	defer tcm.RUnlock()
	ret = tcm.conn

	return
}

// Переподключение к серверу с закрытием прежнего соединения.
func (tcm *tcpConnMng) reconnect() (err error) {
	var conn net.Conn

	tcm.Lock()
	defer tcm.Unlock()
	// Закрытие прежнего соединения.
	if tcm.conn != nil {
		_ = tcm.conn.Close()
		tcm.conn = nil
	}
	// Установка нового соединения с сервером.
	if conn, err = tcm.dial(); err != nil {
		return
	}
	tcm.conn = conn

	return
}

// Write Запись данных в соединение с переподключением при обрыве связи.
// Количество попыток переподключения и пауза между ними задаются политикой переподключения.
func (tcm *tcpConnMng) Write(b []byte) (n int, err error) {
	var (
		attempt int
		conn    net.Conn
	)

	tcm.wmu.Lock()
	defer tcm.wmu.Unlock()
	for {
		// Запись данных в активное соединение.
		if conn = tcm.connection(); conn != nil {
			if n, err = conn.Write(b); err == nil {
				return
			}
		}
		// Исчерпание попыток переподключения.
		if attempt >= tcm.maxReconnect {
			n = 0
			if err == nil {
				err = Errors().NotConnected.Bind()
			}
			return
		}
		attempt++
		// Пауза перед переподключением.
		if tcm.reconnectDelay > 0 {
			time.Sleep(tcm.reconnectDelay)
		}
		// Переподключение к серверу, при ошибке - следующая попытка.
		if err = tcm.reconnect(); err != nil {
			n = 0
			continue
		}
	}
}

// Close Закрытие соединения с сервером.
func (tcm *tcpConnMng) Close() (err error) {
	var conn net.Conn

	tcm.Lock()
	defer tcm.Unlock()
	if conn = tcm.conn; conn == nil {
		return
	}
	tcm.conn = nil
	err = conn.Close()

	return
}
