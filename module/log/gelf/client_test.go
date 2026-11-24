package gelf

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Текст ошибки поддельного клиента сетевого протокола.
const errTextFakeProtocol = "ошибка поддельного клиента сетевого протокола."

// Ошибка поддельного клиента сетевого протокола.
var errFakeProtocol = errors.New(errTextFakeProtocol)

func TestParseAddress(t *testing.T) {
	const (
		tplFatalUnexpectedError = "parseAddress(%q) вернул неожиданную ошибку: %v."
		tplFatalExpectingError  = "parseAddress(%q) должен был вернуть ошибку."
		tplErrorProto           = "parseAddress(%q) вернул протокол %q, ожидался %q."
		tplErrorHost            = "parseAddress(%q) вернул хост %q, ожидался %q."
		tplErrorPort            = "parseAddress(%q) вернул порт %d, ожидался %d."
	)
	var (
		err         error
		proto, host string
		port        uint16
	)

	// Адрес без указания протокола, протокол UDP по умолчанию.
	if proto, host, port, err = parseAddress("backend:12221"); err != nil {
		t.Fatalf(tplFatalUnexpectedError, "backend:12221", err)
	}
	if proto != "udp" || host != "backend" || port != 12221 {
		t.Errorf(tplErrorProto, "backend:12221", proto, "udp")
	}
	// Адрес с протоколом UDP в верхнем регистре.
	if proto, host, port, err = parseAddress("UDP://backend:12221"); err != nil {
		t.Fatalf(tplFatalUnexpectedError, "UDP://backend:12221", err)
	}
	if proto != "udp" || host != "backend" || port != 12221 {
		t.Errorf(tplErrorProto, "UDP://backend:12221", proto, "udp")
	}
	// Адрес с протоколом TCP.
	if proto, host, port, err = parseAddress("tcp://192.168.1.14:12201"); err != nil {
		t.Fatalf(tplFatalUnexpectedError, "tcp://192.168.1.14:12201", err)
	}
	if proto != "tcp" || host != "192.168.1.14" || port != 12201 {
		t.Errorf(tplErrorProto, "tcp://192.168.1.14:12201", proto, "tcp")
	}
	// Адрес с протоколом TLS.
	if proto, host, port, err = parseAddress("tls://backend:12201"); err != nil {
		t.Fatalf(tplFatalUnexpectedError, "tls://backend:12201", err)
	}
	if proto != "tls" || host != "backend" || port != 12201 {
		t.Errorf(tplErrorProto, "tls://backend:12201", proto, "tls")
	}
	// Адрес с протоколом TCP+TLS в верхнем регистре.
	if proto, host, port, err = parseAddress("TCP+TLS://backend:12201"); err != nil {
		t.Fatalf(tplFatalUnexpectedError, "TCP+TLS://backend:12201", err)
	}
	if proto != "tls" || host != "backend" || port != 12201 {
		t.Errorf(tplErrorProto, "TCP+TLS://backend:12201", proto, "tls")
	}
	// Не известный протокол.
	if _, _, _, err = parseAddress("http://backend:12221"); err == nil {
		t.Errorf(tplFatalExpectingError, "http://backend:12221")
	} else if !Errors().UnknownProtocol.Is(err) {
		t.Errorf(tplFatalUnexpectedError, "http://backend:12221", err)
	}
	// Адрес без порта подключения.
	if _, _, _, err = parseAddress("udp://backend"); err == nil {
		t.Errorf(tplFatalExpectingError, "udp://backend")
	} else if !Errors().InvalidAddress.Is(err) {
		t.Errorf(tplFatalUnexpectedError, "udp://backend", err)
	}
	// Адрес с не числовым портом подключения.
	if _, _, _, err = parseAddress("udp://backend:port"); err == nil {
		t.Errorf(tplFatalExpectingError, "udp://backend:port")
	} else if !Errors().InvalidAddress.Is(err) {
		t.Errorf(tplFatalUnexpectedError, "udp://backend:port", err)
	}
	// Адрес с портом подключения за границами допустимого диапазона.
	if _, _, _, err = parseAddress("udp://backend:123456"); err == nil {
		t.Errorf(tplFatalExpectingError, "udp://backend:123456")
	} else if !Errors().InvalidAddress.Is(err) {
		t.Errorf(tplFatalUnexpectedError, "udp://backend:123456", err)
	}
}

func TestNew(t *testing.T) {
	const (
		tplFatalUnexpectedError = "New(%q) вернул неожиданную ошибку: %v."
		tplFatalExpectingError  = "New(%q) должен был вернуть ошибку."
		tplErrorType            = "New(%q) вернул тип %T, ожидался %T."
	)
	var (
		err     error
		lst     net.Listener
		cli     ProtocolClient
		address string
	)

	// Клиент UDP без сжатия.
	if cli, err = New("udp://127.0.0.1:12201", CompressionNone); err != nil {
		t.Fatalf(tplFatalUnexpectedError, "udp://127.0.0.1:12201", err)
	}
	if _, ok := cli.(*client); !ok {
		t.Errorf(tplErrorType, "udp://127.0.0.1:12201", cli, &client{})
	}
	// Клиент UDP со сжатием gzip.
	if cli, err = New("127.0.0.1:12201", CompressionGzip); err != nil {
		t.Fatalf(tplFatalUnexpectedError, "127.0.0.1:12201", err)
	}
	if _, ok := cli.(*client); !ok {
		t.Errorf(tplErrorType, "127.0.0.1:12201", cli, &client{})
	}
	// Клиент UDP со сжатием zlib и настройками клиента.
	if cli, err = New("udp://127.0.0.1:12201", CompressionZlib,
		WithCompressLevel(1), WithChunkSize(512), WithDialTimeout(time.Second)); err != nil {
		t.Fatalf(tplFatalUnexpectedError, "udp://127.0.0.1:12201", err)
	}
	if _, ok := cli.(*client); !ok {
		t.Errorf(tplErrorType, "udp://127.0.0.1:12201", cli, &client{})
	}
	// Не известный тип сжатия.
	if _, err = New("udp://127.0.0.1:12201", Compression(42)); err == nil {
		t.Errorf(tplFatalExpectingError, "udp://127.0.0.1:12201")
	} else if !Errors().UnknownCompression.Is(err) {
		t.Errorf(tplFatalUnexpectedError, "udp://127.0.0.1:12201", err)
	}
	// Ошибка разбора адреса сервера: порт подключения за границами допустимого диапазона.
	if _, err = New("udp://backend:123456", CompressionNone); err == nil {
		t.Errorf(tplFatalExpectingError, "udp://backend:123456")
	} else if !Errors().InvalidAddress.Is(err) {
		t.Errorf(tplFatalUnexpectedError, "udp://backend:123456", err)
	}
	// Ошибка создания клиента UDP с недопустимым символом в имени хоста сервера.
	if _, err = New("udp://ho]st:80", CompressionNone); err == nil {
		t.Errorf(tplFatalExpectingError, "udp://ho]st:80")
	}
	// Ошибка создания клиента TCP с недопустимым символом в имени хоста сервера.
	if _, err = New("tcp://ho]st:80", CompressionNone); err == nil {
		t.Errorf(tplFatalExpectingError, "tcp://ho]st:80")
	}
	// Протоколы TCP и TLS не поддерживают сжатие сообщений.
	for _, address = range []string{"tcp://127.0.0.1:12201", "tls://127.0.0.1:12201"} {
		if _, err = New(address, CompressionGzip); err == nil {
			t.Errorf(tplFatalExpectingError, address)
		} else if !Errors().CompressionNotSupported.Is(err) {
			t.Errorf(tplFatalUnexpectedError, address, err)
		}
		if _, err = New(address, CompressionZlib); err == nil {
			t.Errorf(tplFatalExpectingError, address)
		} else if !Errors().CompressionNotSupported.Is(err) {
			t.Errorf(tplFatalUnexpectedError, address, err)
		}
	}
	// Ошибка подключения клиента TCP к закрытому порту сервера.
	if lst, err = net.Listen(tcpNetwork, "127.0.0.1:0"); err != nil {
		t.Fatalf(tplFatalUnexpectedError, "net.Listen()", err)
	}
	address = lst.Addr().String()
	_ = lst.Close()
	if _, err = New("tcp://"+address, CompressionNone, WithReconnect(1, time.Millisecond)); err == nil {
		t.Errorf(tplFatalExpectingError, "tcp://"+address)
	}
	// TCP сервер для проверки клиентов TCP.
	if lst, err = net.Listen(tcpNetwork, "127.0.0.1:0"); err != nil {
		t.Fatalf(tplFatalUnexpectedError, "net.Listen()", err)
	}
	defer func() { _ = lst.Close() }()
	address = lst.Addr().String()
	// Клиент TCP без сжатия.
	if cli, err = New("tcp://"+address, CompressionNone); err != nil {
		t.Fatalf(tplFatalUnexpectedError, "tcp://"+address, err)
	}
	if _, ok := cli.(*client); !ok {
		t.Errorf(tplErrorType, "tcp://"+address, cli, &client{})
	}
}

func TestClientSend(t *testing.T) {
	const (
		tplFatalListenError     = "net.ListenUDP() вернул неожиданную ошибку: %v."
		tplFatalUnexpectedError = "вызов функции вернул неожиданную ошибку: %v."
		tplFatalExpectingError  = "Send() должен был вернуть ошибку."
		tplErrorMessage         = "получено неверное сообщение: %q, ожидалось %q."
	)
	var (
		err      error
		lst      *net.UDPConn
		cli      Client
		msg      *Message
		expected []byte
		chunk    []byte
		data     []byte
		length   int
		reader   io.ReadCloser
	)

	// UDP сервер для приёма сообщений.
	if lst, err = net.ListenUDP(udpNetwork, &net.UDPAddr{IP: net.ParseIP("127.0.0.1")}); err != nil {
		t.Fatalf(tplFatalListenError, err)
	}
	defer func() { _ = lst.Close() }()
	// Отправка сообщения без сжатия.
	if cli, err = New("udp://"+lst.LocalAddr().String(), CompressionNone); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	msg = NewMessage("test-host", "test message", SeverityInfo)
	msg.Keys["key1"] = "value1"
	if expected, err = json.Marshal(msg); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if err = cli.Send(msg); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	chunk = make([]byte, 4096)
	if length, err = udpRead(lst, chunk); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if !bytes.Equal(chunk[chunkHeaderSize:length], expected) {
		t.Errorf(tplErrorMessage, chunk[chunkHeaderSize:length], expected)
	}
	// Отправка сообщения со сжатием gzip.
	if cli, err = New("udp://"+lst.LocalAddr().String(), CompressionGzip); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if err = cli.Send(msg); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if length, err = udpRead(lst, chunk); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if reader, err = gzip.NewReader(bytes.NewReader(chunk[chunkHeaderSize:length])); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if data, err = io.ReadAll(reader); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if !bytes.Equal(data, expected) {
		t.Errorf(tplErrorMessage, data, expected)
	}
	// Отправка сообщения со сжатием zlib.
	if cli, err = New("udp://"+lst.LocalAddr().String(), CompressionZlib); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if err = cli.Send(msg); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if length, err = udpRead(lst, chunk); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if reader, err = zlib.NewReader(bytes.NewReader(chunk[chunkHeaderSize:length])); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if data, err = io.ReadAll(reader); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if !bytes.Equal(data, expected) {
		t.Errorf(tplErrorMessage, data, expected)
	}
	// Ошибка кодирования сообщения с дополнительным полем недопустимого типа.
	msg.Keys["fn"] = func() {}
	if err = cli.Send(msg); err == nil {
		t.Errorf(tplFatalExpectingError)
	}
	// Ошибка сжатия сообщения с недопустимым уровнем сжатия gzip и zlib.
	cli = &client{compression: CompressionGzip, compressLevel: 100}
	if err = cli.Send(NewMessage("test-host", "test message", SeverityInfo)); err == nil {
		t.Errorf(tplFatalExpectingError)
	}
	cli = &client{compression: CompressionZlib, compressLevel: 100}
	if err = cli.Send(NewMessage("test-host", "test message", SeverityInfo)); err == nil {
		t.Errorf(tplFatalExpectingError)
	}
	// Ошибка отправки данных сетевым протоколом.
	cli = &client{protocol: &fakeProtocolClient{err: errFakeProtocol}, compression: CompressionNone}
	if err = cli.Send(NewMessage("test-host", "test message", SeverityInfo)); !errors.Is(err, errFakeProtocol) {
		t.Errorf(tplFatalExpectingError)
	}
}

func TestClientSendTcp(t *testing.T) {
	const (
		tplFatalListenError     = "net.Listen() вернул неожиданную ошибку: %v."
		tplFatalUnexpectedError = "вызов функции вернул неожиданную ошибку: %v."
		tplErrorMessage         = "получено неверное сообщение: %q, ожидалось %q."
	)
	var (
		err      error
		lst      net.Listener
		cli      Client
		msg      *Message
		expected []byte
		conn     net.Conn
		message  []byte
	)

	// TCP сервер для приёма сообщений.
	if lst, err = net.Listen(tcpNetwork, "127.0.0.1:0"); err != nil {
		t.Fatalf(tplFatalListenError, err)
	}
	defer func() { _ = lst.Close() }()
	if cli, err = New("tcp://"+lst.Addr().String(), CompressionNone); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	defer func() { _ = cli.Close() }()
	// Отправка сообщения с нулевым байтом-разделителем.
	msg = NewMessage("test-host", "test message", SeverityInfo)
	if expected, err = json.Marshal(msg); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if err = cli.Send(msg); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	// Приём и проверка сообщения с байтом-разделителем.
	if conn, err = lst.Accept(); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	defer func() { _ = conn.Close() }()
	if message, err = bufio.NewReader(conn).ReadBytes(0); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	expected = append(expected, 0)
	if !bytes.Equal(message, expected) {
		t.Errorf(tplErrorMessage, message, expected)
	}
}

func TestClientWrite(t *testing.T) {
	const (
		tplFatalListenError     = "net.ListenUDP() вернул неожиданную ошибку: %v."
		tplFatalUnexpectedError = "вызов функции вернул неожиданную ошибку: %v."
		tplErrorWriteCount      = "Write() вернул количество байт %d, ожидалось %d."
		tplErrorFieldValue      = "поле %q содержит неверное значение: %v."
		tplErrorFieldNotFound   = "поле %q не найдено в сообщении."
	)
	var (
		err      error
		lst      *net.UDPConn
		cli      Client
		hostname string
		text     []byte
		chunk    []byte
		length   int
		item     map[string]any
	)

	// UDP сервер для приёма сообщений.
	if lst, err = net.ListenUDP(udpNetwork, &net.UDPAddr{IP: net.ParseIP("127.0.0.1")}); err != nil {
		t.Fatalf(tplFatalListenError, err)
	}
	defer func() { _ = lst.Close() }()
	if cli, err = New("udp://"+lst.LocalAddr().String(), CompressionNone); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if hostname, err = os.Hostname(); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	// Отправка однострочного текста через интерфейс io.Writer.
	text = []byte("single line message")
	if length, err = cli.Write(text); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if length != len(text) {
		t.Errorf(tplErrorWriteCount, length, len(text))
	}
	chunk = make([]byte, 4096)
	if length, err = udpRead(lst, chunk); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if err = json.Unmarshal(chunk[chunkHeaderSize:length], &item); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if item["host"] != hostname {
		t.Errorf(tplErrorFieldValue, "host", item["host"])
	}
	if item["short_message"] != "single line message" {
		t.Errorf(tplErrorFieldValue, "short_message", item["short_message"])
	}
	if item["level"] != float64(SeverityInfo) {
		t.Errorf(tplErrorFieldValue, "level", item["level"])
	}
	if item["facility"] != filepath.Base(os.Args[0]) {
		t.Errorf(tplErrorFieldValue, "facility", item["facility"])
	}
	if _, ok := item["full_message"]; ok {
		t.Errorf(tplErrorFieldValue, "full_message", item["full_message"])
	}
	// Отправка многострочного текста: первая строка - краткий текст, весь текст - полный.
	item = nil
	text = []byte("short line\nfull message text")
	if _, err = cli.Write(text); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if length, err = udpRead(lst, chunk); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if err = json.Unmarshal(chunk[chunkHeaderSize:length], &item); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if item["short_message"] != "short line" {
		t.Errorf(tplErrorFieldValue, "short_message", item["short_message"])
	}
	if item["full_message"] != "short line\nfull message text" {
		t.Errorf(tplErrorFieldValue, "full_message", item["full_message"])
	}
	// Ошибка отправки сообщения сетевым протоколом.
	cli = &client{protocol: &fakeProtocolClient{err: errFakeProtocol}}
	if _, err = cli.Write(text); !errors.Is(err, errFakeProtocol) {
		t.Errorf(tplErrorFieldNotFound, "ошибка протокола")
	}
}

func TestClientSendData(t *testing.T) {
	const (
		tplFatalUnexpectedError = "SendData() вернул неожиданную ошибку: %v."
		tplFatalExpectingError  = "SendData() должен был вернуть ошибку."
		tplErrorData            = "клиент протокола получил неверные данные: %q."
	)
	var (
		err  error
		fake *fakeProtocolClient
		cli  *client
	)

	fake = &fakeProtocolClient{}
	cli = &client{protocol: fake}
	// Отправка подготовленных данных сообщения.
	if err = cli.SendData([]byte("data")); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if !bytes.Equal(fake.data, []byte("data")) {
		t.Errorf(tplErrorData, fake.data)
	}
	// Ошибка отправки данных клиентом сетевого протокола.
	fake.err = errFakeProtocol
	if err = cli.SendData([]byte("data")); !errors.Is(err, errFakeProtocol) {
		t.Errorf(tplFatalExpectingError)
	}
}

func TestClientClose(t *testing.T) {
	const (
		tplFatalUnexpectedError = "Close() вернул неожиданную ошибку: %v."
		tplFatalExpectingError  = "Close() должен был вернуть ошибку."
	)
	var (
		err  error
		fake *fakeProtocolClient
		cli  *client
	)

	fake = &fakeProtocolClient{}
	cli = &client{protocol: fake}
	// Закрытие соединения клиентом сетевого протокола.
	if err = cli.Close(); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	// Ошибка закрытия соединения клиентом сетевого протокола.
	fake.err = errFakeProtocol
	if err = cli.Close(); !errors.Is(err, errFakeProtocol) {
		t.Errorf(tplFatalExpectingError)
	}
}

// Чтение одной датаграммы из UDP-соединения с ограничением времени ожидания.
func udpRead(conn *net.UDPConn, buf []byte) (n int, err error) {
	if err = conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		return
	}
	if n, _, err = conn.ReadFromUDP(buf); err != nil {
		return
	}

	return
}

// Поддельный клиент сетевого протокола для проверки обработки ошибок отправки данных.
type fakeProtocolClient struct {
	data []byte // Полученные данные сообщения.
	err  error  // Ошибка, возвращаемая методами клиента.
}

// SendData Приём данных сообщения с возвратом заданной ошибки.
func (fpc *fakeProtocolClient) SendData(data []byte) (err error) {
	fpc.data = data
	err = fpc.err

	return
}

// Close Возврат заданной ошибки закрытия соединения.
func (fpc *fakeProtocolClient) Close() (err error) {
	err = fpc.err

	return
}
