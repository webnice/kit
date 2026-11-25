package log

import (
	"bytes"
	"encoding/json"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	kitModuleLogGelf "github.com/webnice/kit/v5/module/log/gelf"
	kitModuleLogLevel "github.com/webnice/kit/v5/module/log/level"
	kitTypes "github.com/webnice/kit/v5/types"
)

// Название приложения, передаваемое обработчику Graylog в тестах.
const testGraylogAppName = "test-app"

// Приём одного сообщения GELF, отправленного обработчиком Graylog, и возврат разобранных полей сообщения.
func graylogReceive(t *testing.T, msg *Message) (item map[string]any) {
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
	glg = &graylog{address: address, hostname: testHostname, appName: testGraylogAppName, cli: cli, wr: new(testSyncWriter)}
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
		item  map[string]any
		short string
	)

	// Не форматный вызов логирования: шаблон пустой, аргументы склеиваются без форматирования.
	msg = &Message{Pattern: &bytes.Buffer{}, Argument: []any{testArgument}, Keys: make(map[string]any)}
	item = graylogReceive(t, msg)
	short, _ = item["short_message"].(string)
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
		item  map[string]any
		short string
	)

	// Форматный вызов логирования: аргументы подставляются в шаблон.
	msg = &Message{Pattern: bytes.NewBufferString(testPattern), Argument: []any{500}, Keys: make(map[string]any)}
	item = graylogReceive(t, msg)
	short, _ = item["short_message"].(string)
	if short != testExpected {
		t.Errorf(tplMismatch, short, testExpected)
	}
}

func TestGraylogHandlerMessageFields(t *testing.T) {
	const (
		tplMismatch = "получено значение поля %q равное %v, ожидалось %v."
		tplEmpty    = "поле %q отсутствует или пустое."
		testPattern = "код ответа: %d"
		testMessage = "код ответа: 500"
		testHost    = "test-host"
	)
	var (
		msg      *Message
		item     map[string]any
		expected map[string]any
		trace    *kitTypes.TraceInfo
		key      string
	)

	trace = kitTypes.NewTraceInfo()
	trace.Package.WriteString("github.com/webnice/kit/v5/module/log")
	trace.FilenameLong.WriteString("/a/b/log/graylog.go")
	trace.FilenameShort.WriteString("graylog.go")
	trace.Function.WriteString("Handler")
	trace.StackTrace.WriteString("стек вызовов")
	trace.Line = 42
	msg = &Message{
		Timestamp: time.Now(),
		Level:     kitModuleLogLevel.Warning,
		Pattern:   bytes.NewBufferString(testPattern),
		Argument:  []any{500},
		Keys:      make(map[string]any),
		Trace:     trace,
	}
	item = graylogReceive(t, msg)
	expected = map[string]any{
		"short_message": testMessage,
		"full_message":  testMessage,
		"message":       testMessage,
		"facility":      "github.com/webnice/kit/v5/module/log",
		"package":       "github.com/webnice/kit/v5/module/log",
		"module":        "log",
		"filenameLong":  "/a/b/log/graylog.go",
		"filenameShort": "graylog.go",
		"fileLine":      float64(42),
		"function":      "Handler",
		"callStack":     "стек вызовов",
		"levelString":   "warning",
		"pid":           float64(os.Getpid()),
		"hostName":      testHost,
		"appName":       testGraylogAppName,
		"_file":         "graylog.go",
		"_line":         float64(42),
		"_function":     "Handler",
	}
	for key = range expected {
		if item[key] != expected[key] {
			t.Errorf(tplMismatch, key, item[key], expected[key])
		}
	}
	// Идентификатор сообщения и локальное время сообщения присутствуют и не пустые.
	if id, _ := item["id"].(string); id == "" {
		t.Errorf(tplEmpty, "id")
	}
	if now, _ := item["todayAndNow"].(string); now == "" {
		t.Errorf(tplEmpty, "todayAndNow")
	}
}
