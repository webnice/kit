package log

import (
	"bytes"
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"

	kitModuleLogGelf "github.com/webnice/kit/v5/module/log/gelf"
)

// Приём одного сообщения GELF, отправленного обработчиком Graylog, и возврат значения поля short_message.
func graylogReceiveShort(t *testing.T, msg *Message) (short string) {
	const (
		tplFatalListen     = "функция net.ListenUDP(), вернула не ожидаемую ошибку: %s."
		tplFatalDeadline   = "функция SetReadDeadline(), вернула не ожидаемую ошибку: %s."
		tplFatalClient     = "функция gelf.New(), вернула не ожидаемую ошибку: %s."
		tplFatalRead       = "функция ReadFromUDP(), вернула не ожидаемую ошибку: %s."
		tplFatalShortChunk = "получен чанк короче заголовка GELF UDP: %d байт."
		tplFatalUnmarshal  = "функция json.Unmarshal(), вернула не ожидаемую ошибку: %s."
		protoUdp           = "udp"
		testHostname       = "test-host"
		readTimeout        = time.Second * 3
		// Размер заголовка чанка GELF UDP: магический заголовок, идентификатор сообщения, последовательность.
		chunkHeaderSize = 2 + 8 + 2
	)
	var (
		lst     *net.UDPConn
		cli     kitModuleLogGelf.Client
		glg     *graylog
		buf     []byte
		item    map[string]any
		address string
		n       int
		err     error
	)

	if lst, err = net.ListenUDP(protoUdp, &net.UDPAddr{IP: net.ParseIP("127.0.0.1")}); err != nil {
		t.Fatalf(tplFatalListen, err)
	}
	defer func() { _ = lst.Close() }()
	address = protoUdp + "://" + lst.LocalAddr().String()
	if cli, err = kitModuleLogGelf.New(address, kitModuleLogGelf.CompressionNone); err != nil {
		t.Fatalf(tplFatalClient, err)
	}
	glg = &graylog{address: address, hostname: testHostname, cli: cli, wr: new(testSyncWriter)}
	glg.Handler(msg)
	buf = make([]byte, 65535)
	if err = lst.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		t.Fatalf(tplFatalDeadline, err)
	}
	if n, _, err = lst.ReadFromUDP(buf); err != nil {
		t.Fatalf(tplFatalRead, err)
	}
	if n <= chunkHeaderSize {
		t.Fatalf(tplFatalShortChunk, n)
	}
	item = make(map[string]any)
	if err = json.Unmarshal(buf[chunkHeaderSize:n], &item); err != nil {
		t.Fatalf(tplFatalUnmarshal, err)
	}
	short, _ = item["short_message"].(string)

	return
}

func TestGraylogHandlerEmptyPattern(t *testing.T) {
	const (
		tplMismatch  = "получено сообщение %q, ожидалось %q."
		tplExtra     = "в сообщении %q присутствует признак ошибки форматирования: %s."
		testArgument = "[ 25.11.2026 17:50:34:0000017681 ] ------------------------------------------------------------"
		testExtra    = "%!(EXTRA"
	)
	var (
		msg   *Message
		short string
	)

	// Не форматный вызов логирования: шаблон пустой, аргументы склеиваются без форматирования.
	msg = &Message{Pattern: &bytes.Buffer{}, Argument: []any{testArgument}, Keys: make(map[string]any)}
	short = graylogReceiveShort(t, msg)
	if short != testArgument {
		t.Errorf(tplMismatch, short, testArgument)
	}
	if strings.Contains(short, testExtra) {
		t.Errorf(tplExtra, short, testExtra)
	}
}

func TestGraylogHandlerPatternWithArguments(t *testing.T) {
	const (
		tplMismatch  = "получено сообщение %q, ожидалось %q."
		testPattern  = "код ответа: %d"
		testExpected = "код ответа: 500"
	)
	var (
		msg   *Message
		short string
	)

	// Форматный вызов логирования: аргументы подставляются в шаблон.
	msg = &Message{Pattern: bytes.NewBufferString(testPattern), Argument: []any{500}, Keys: make(map[string]any)}
	short = graylogReceiveShort(t, msg)
	if short != testExpected {
		t.Errorf(tplMismatch, short, testExpected)
	}
}
