package gelf

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"io"
)

// Send Кодирование и отправка сообщения серверу GELF.
func (cli *client) Send(msg *Message) (err error) {
	var data []byte

	// Кодирование сообщения в формат JSON.
	if data, err = json.Marshal(msg); err != nil {
		return
	}
	// Сжатие данных сообщения.
	switch cli.compression {
	case CompressionGzip, CompressionZlib:
		if data, err = cli.compressData(data); err != nil {
			return
		}
	case CompressionNone:
	}
	// Отправка данных серверу.
	if err = cli.protocol.SendData(data); err != nil {
		return
	}

	return
}

// Write Реализация интерфейса io.Writer: отправка строки текста сообщением
// с уровнем важности SeverityInfo. Первая строка текста становится кратким
// текстом сообщения, весь текст целиком - полным текстом сообщения.
func (cli *client) Write(p []byte) (n int, err error) {
	var (
		msg   *Message
		short []byte
		full  []byte
		i     int
	)

	n = len(p)
	p = bytes.TrimSpace(p)
	short = p
	// Первая строка многострочного текста - краткий текст, весь текст - полный.
	if i = bytes.IndexByte(p, '\n'); i > 0 {
		short, full = p[:i], p
	}
	// Формирование и отправка сообщения.
	msg = NewMessage(cli.hostname, string(short), SeverityInfo)
	msg.Full = string(full)
	msg.Facility = cli.facility
	if err = cli.Send(msg); err != nil {
		n = 0
		return
	}

	return
}

// Сжатие данных сообщения алгоритмом gzip или zlib.
func (cli *client) compressData(data []byte) (ret []byte, err error) {
	var (
		buf bytes.Buffer
		zwr io.WriteCloser
	)

	// Создание потока сжатия.
	switch cli.compression {
	case CompressionGzip:
		if zwr, err = gzip.NewWriterLevel(&buf, cli.compressLevel); err != nil {
			return
		}
	case CompressionZlib:
		if zwr, err = zlib.NewWriterLevel(&buf, cli.compressLevel); err != nil {
			return
		}
	}
	// Сжатие данных сообщения.
	if _, err = zwr.Write(data); err != nil {
		return
	}
	if err = zwr.Close(); err != nil {
		return
	}
	ret = buf.Bytes()

	return
}

// SendData Отправка подготовленных данных сообщения серверу.
func (cli *client) SendData(data []byte) (err error) {
	err = cli.protocol.SendData(data)

	return
}

// Close Закрытие соединения и освобождение занятых ресурсов.
func (cli *client) Close() (err error) {
	err = cli.protocol.Close()

	return
}
