package gelf

import (
	"crypto/rand"
	"io"
	"net"
	"strconv"
)

// NewUdp Создание клиента протокола UDP.
func NewUdp(host string, port uint16, chunkSize uint, opts ...Option) (ret ProtocolClient, err error) {
	var (
		opt   options
		apply Option
	)

	// Применение настроек клиента.
	opt = defaultOptions()
	for _, apply = range opts {
		apply(&opt)
	}
	if chunkSize != 0 {
		opt.chunkSize = chunkSize
	}
	// Создание клиента протокола UDP.
	if ret, err = newUdpClient(host, port, opt); err != nil {
		return
	}

	return
}

// Создание клиента протокола UDP на основе настроек и установка постоянного соединения с сервером.
func newUdpClient(host string, port uint16, opt options) (ret ProtocolClient, err error) {
	var (
		dialer  net.Dialer
		address string
		conn    net.Conn
	)

	// Установка постоянного соединения с сервером.
	address = net.JoinHostPort(host, strconv.FormatUint(uint64(port), 10))
	dialer.Timeout = opt.dialTimeout
	if conn, err = dialer.Dial(udpNetwork, address); err != nil {
		return
	}
	ret = &udpClient{
		conn:      conn.(*net.UDPConn),
		chunkSize: opt.chunkSize,
	}

	return
}

// SendData Отправка данных сообщения с разбивкой на чанки.
func (cli *udpClient) SendData(data []byte) (err error) {
	var (
		messageID  []byte
		chunkCount int
		begin, end int
		n          int
	)

	// Вычисление количества чанков сообщения.
	chunkCount = len(data) / int(cli.chunkSize)
	if len(data)%int(cli.chunkSize) > 0 {
		chunkCount++
	}
	if chunkCount > maxChunkCount {
		err = Errors().MessageTooLarge.Bind(len(data), int(cli.chunkSize)*maxChunkCount)
		return
	}
	// Создание идентификатора сообщения и отправка чанков.
	if messageID, err = newMessageID(); err != nil {
		return
	}
	for n = 0; n < chunkCount; n++ {
		begin = n * int(cli.chunkSize)
		end = begin + int(cli.chunkSize)
		if end > len(data) {
			end = len(data)
		}
		if err = cli.sendChunk(messageID, byte(n), byte(chunkCount), data[begin:end]); err != nil {
			return
		}
	}

	return
}

// Отправка одного чанка сообщения.
func (cli *udpClient) sendChunk(messageID []byte, index, count byte, chunk []byte) (err error) {
	var (
		data []byte
		n    int
	)

	// Формирование чанка сообщения с заголовком.
	data = make([]byte, 0, chunkHeaderSize+len(chunk))
	data = append(data, chunkMagic0, chunkMagic1)
	data = append(data, messageID...)
	data = append(data, index, count)
	data = append(data, chunk...)
	// Отправка чанка сообщения.
	if n, err = cli.conn.Write(data); err != nil {
		return
	}
	if n != len(data) {
		err = io.ErrShortWrite
		return
	}

	return
}

// Создание уникального идентификатора сообщения.
func newMessageID() (ret []byte, err error) {
	ret = make([]byte, messageIDSize)
	if _, err = rand.Read(ret); err != nil {
		return
	}

	return
}

// Close Закрытие соединения с сервером.
func (cli *udpClient) Close() (err error) {
	err = cli.conn.Close()

	return
}
