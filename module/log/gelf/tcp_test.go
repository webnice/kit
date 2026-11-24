package gelf

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"math/big"
	"net"
	"testing"
	"time"

	wns "github.com/webnice/net"
)

func TestNewTcp(t *testing.T) {
	const (
		tplFatalListenError     = "net.Listen() вернул неожиданную ошибку: %v."
		tplFatalUnexpectedError = "NewTcp() вернул неожиданную ошибку: %v."
		tplFatalExpectingError  = "NewTcp() должен был вернуть ошибку."
	)
	var (
		err  error
		lst  net.Listener
		cli  ProtocolClient
		host string
		port uint16
	)

	// Ошибка разбора адреса с недопустимым символом в имени хоста.
	if _, err = NewTcp("ho]st", 80); err == nil {
		t.Errorf(tplFatalExpectingError)
	}
	// Ошибка подключения к закрытому порту сервера.
	if lst, err = net.Listen(tcpNetwork, "127.0.0.1:0"); err != nil {
		t.Fatalf(tplFatalListenError, err)
	}
	host = lst.Addr().(*net.TCPAddr).IP.String()
	port = uint16(lst.Addr().(*net.TCPAddr).Port)
	_ = lst.Close()
	if _, err = NewTcp(host, port, WithReconnect(1, time.Millisecond)); err == nil {
		t.Errorf(tplFatalExpectingError)
	}
	// Успешное создание клиента протокола TCP.
	if lst, err = net.Listen(tcpNetwork, "127.0.0.1:0"); err != nil {
		t.Fatalf(tplFatalListenError, err)
	}
	defer func() { _ = lst.Close() }()
	host = lst.Addr().(*net.TCPAddr).IP.String()
	port = uint16(lst.Addr().(*net.TCPAddr).Port)
	if cli, err = NewTcp(host, port); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if err = cli.Close(); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
}

func TestTcpClientSendData(t *testing.T) {
	const (
		tplFatalListenError     = "net.Listen() вернул неожиданную ошибку: %v."
		tplFatalUnexpectedError = "SendData() вернул неожиданную ошибку: %v."
		tplErrorMessage         = "получено неверное сообщение: %q."
	)
	var (
		err     error
		lst     net.Listener
		cli     ProtocolClient
		host    string
		port    uint16
		data    []byte
		receive chan []byte
	)

	if lst, err = net.Listen(tcpNetwork, "127.0.0.1:0"); err != nil {
		t.Fatalf(tplFatalListenError, err)
	}
	defer func() { _ = lst.Close() }()
	if host, _, err = net.SplitHostPort(lst.Addr().String()); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	port = uint16(lst.Addr().(*net.TCPAddr).Port)
	if cli, err = NewTcp(host, port); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	defer func() { _ = cli.Close() }()
	// Приём сообщения в отдельной горутине.
	receive = make(chan []byte, 1)
	go func() {
		var (
			conn   net.Conn
			accept error
			buf    []byte
			n      int
		)

		if conn, accept = lst.Accept(); accept != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		buf = make([]byte, 4096)
		if n, accept = conn.Read(buf); accept != nil {
			return
		}
		receive <- buf[:n]
	}()
	// Отправка сообщения с разделителем.
	data = []byte(`{"version":"1.1","short_message":"test"}`)
	if err = cli.SendData(data); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	// Проверка полученного сообщения.
	var (
		message []byte
		wait    <-chan time.Time
	)

	wait = time.After(time.Second)
	select {
	case message = <-receive:
		if !bytes.Equal(message[:len(message)-1], data) || message[len(message)-1] != 0 {
			t.Errorf(tplErrorMessage, message)
		}
	case <-wait:
		t.Fatalf(tplErrorMessage, "timeout")
	}
}

func TestTcpClientSendDataError(t *testing.T) {
	const (
		tplFatalListenError     = "net.Listen() вернул неожиданную ошибку: %v."
		tplFatalUnexpectedError = "вызов функции вернул неожиданную ошибку: %v."
		tplFatalExpectingError  = "SendData() должен был вернуть ошибку."
	)
	var (
		err  error
		lst  net.Listener
		cli  ProtocolClient
		host string
		port uint16
	)

	if lst, err = net.Listen(tcpNetwork, "127.0.0.1:0"); err != nil {
		t.Fatalf(tplFatalListenError, err)
	}
	if host, _, err = net.SplitHostPort(lst.Addr().String()); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	port = uint16(lst.Addr().(*net.TCPAddr).Port)
	if cli, err = NewTcp(host, port, WithReconnect(2, time.Millisecond)); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	// Разрыв соединения с остановкой сервера: переподключение невозможно.
	_ = lst.Close()
	_ = cli.Close()
	if err = cli.SendData([]byte("data")); err == nil {
		t.Errorf(tplFatalExpectingError)
	}
}

func TestTcpConnMngWrite(t *testing.T) {
	const (
		tplFatalListenError     = "net.Listen() вернул неожиданную ошибку: %v."
		tplFatalUnexpectedError = "вызов функции вернул неожиданную ошибку: %v."
		tplFatalExpectingError  = "Write() должен был вернуть ошибку."
		tplErrorWriteCount      = "Write() записал %d байт, ожидалось %d."
		tplErrorMessage         = "получено неверное сообщение: %q, ожидалось %q."
		tplErrorDialCount       = "выполнено %d попыток подключения, ожидалось %d."
	)
	var (
		err     error
		lst     net.Listener
		addr    *net.TCPAddr
		tcm     *tcpConnMng
		conn    net.Conn
		data    []byte
		message []byte
		dials   int
		n       int
	)

	data = []byte("test data")
	if lst, err = net.Listen(tcpNetwork, "127.0.0.1:0"); err != nil {
		t.Fatalf(tplFatalListenError, err)
	}
	defer func() { _ = lst.Close() }()
	addr = lst.Addr().(*net.TCPAddr)
	// Запись без активного соединения: автоматическое подключение к серверу.
	tcm = &tcpConnMng{
		dial:           tcpDialFunc(addr.String()),
		maxReconnect:   3,
		reconnectDelay: time.Millisecond,
	}
	if n, err = tcm.Write(data); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if n != len(data) {
		t.Errorf(tplErrorWriteCount, n, len(data))
	}
	if message, err = tcpReceive(lst); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if !bytes.Equal(message, data) {
		t.Errorf(tplErrorMessage, message, data)
	}
	_ = tcm.Close()
	// Запись после разрыва соединения: переподключение и повторная отправка данных.
	tcm = &tcpConnMng{
		dial:           tcpDialFunc(addr.String()),
		maxReconnect:   3,
		reconnectDelay: time.Millisecond,
	}
	if err = tcm.reconnect(); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	// Разрыв активного соединения без ведома менеджера соединения.
	if conn = tcm.connection(); conn != nil {
		_ = conn.Close()
	}
	if n, err = tcm.Write(data); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if n != len(data) {
		t.Errorf(tplErrorWriteCount, n, len(data))
	}
	// Первое соединение закрыто клиентом без данных, данные приходят по второму соединению.
	if err = tcpDiscard(lst); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if message, err = tcpReceive(lst); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if !bytes.Equal(message, data) {
		t.Errorf(tplErrorMessage, message, data)
	}
	_ = tcm.Close()
	// Ошибка записи с невозможностью переподключения: сервер остановлен.
	tcm = &tcpConnMng{
		dial:           tcpDialFunc(addr.String()),
		maxReconnect:   2,
		reconnectDelay: time.Millisecond,
	}
	if err = tcm.reconnect(); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	_ = lst.Close()
	if conn = tcm.connection(); conn != nil {
		_ = conn.Close()
	}
	if _, err = tcm.Write(data); err == nil {
		t.Errorf(tplFatalExpectingError)
	}
	// Исчерпание попыток переподключения без активного соединения: две попытки подключения.
	dials = 0
	tcm = &tcpConnMng{
		dial: func() (net.Conn, error) {
			dials++
			return tcpDialFunc(addr.String())()
		},
		maxReconnect:   2,
		reconnectDelay: 0,
	}
	if _, err = tcm.Write(data); err == nil {
		t.Errorf(tplFatalExpectingError)
	}
	if dials != 2 {
		t.Errorf(tplErrorDialCount, dials, 2)
	}
	// Отсутствие активного соединения при нулевом количестве попыток переподключения.
	tcm = &tcpConnMng{dial: tcpDialFunc(addr.String())}
	if _, err = tcm.Write(data); err == nil {
		t.Errorf(tplFatalExpectingError)
	} else if !Errors().NotConnected.Is(err) {
		t.Errorf(tplFatalUnexpectedError, err)
	}
}

func TestTcpConnMngClose(t *testing.T) {
	const (
		tplFatalUnexpectedError = "Close() вернул неожиданную ошибку: %v."
	)
	var (
		err error
		tcm *tcpConnMng
	)

	// Закрытие менеджера соединения без активного соединения.
	tcm = &tcpConnMng{}
	if err = tcm.Close(); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
}

func TestNewTcpTls(t *testing.T) {
	const (
		tplFatalListenError     = "net.Listen() вернул неожиданную ошибку: %v."
		tplFatalUnexpectedError = "вызов функции вернул неожиданную ошибку: %v."
		tplFatalExpectingError  = "New() должен был вернуть ошибку."
		tplErrorMessage         = "получено неверное сообщение: %q, ожидалось %q."
	)
	var (
		err      error
		lst      net.Listener
		tlsLst   net.Listener
		srv      wns.Interface
		srvCfg   *tls.Config
		cliCfg   *tls.Config
		cli      Client
		msg      *Message
		expected []byte
		receive  chan []byte
		message  []byte
		wait     <-chan time.Time
	)

	// Самоподписанный сертификат и конфигурации TLS сервера и клиента.
	if srvCfg, cliCfg, err = newTestTLSConfigs(); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	// TLS сервер на базе библиотеки webnice/net с прослушиванием переданного слушателя.
	if lst, err = net.Listen(tcpNetwork, "127.0.0.1:0"); err != nil {
		t.Fatalf(tplFatalListenError, err)
	}
	tlsLst = tls.NewListener(lst, srvCfg)
	receive = make(chan []byte, 1)
	srv = wns.New().Handler(func(l net.Listener) error {
		var (
			conn net.Conn
			buf  []byte
			berr error
		)

		if conn, berr = l.Accept(); berr != nil {
			return berr
		}
		defer func() { _ = conn.Close() }()
		if buf, berr = bufio.NewReader(conn).ReadBytes(0); berr != nil {
			return berr
		}
		receive <- buf

		return nil
	})
	srv.Serve(tlsLst)
	defer srv.Stop()
	// Отправка сообщения по протоколу TLS.
	if cli, err = New("tls://"+lst.Addr().String(), CompressionNone, WithTLSConfig(cliCfg)); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	defer func() { _ = cli.Close() }()
	msg = NewMessage("test-host", "test message", SeverityInfo)
	if expected, err = json.Marshal(msg); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if err = cli.Send(msg); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	// Проверка полученного сообщения с байтом-разделителем.
	wait = time.After(time.Second)
	select {
	case message = <-receive:
		expected = append(expected, 0)
		if !bytes.Equal(message, expected) {
			t.Errorf(tplErrorMessage, message, expected)
		}
	case <-wait:
		t.Fatalf(tplErrorMessage, "timeout", expected)
	}
	// Ошибка проверки подлинности сервера с конфигурацией TLS по умолчанию.
	if lst, err = net.Listen(tcpNetwork, "127.0.0.1:0"); err != nil {
		t.Fatalf(tplFatalListenError, err)
	}
	tlsLst = tls.NewListener(lst, srvCfg)
	srv = wns.New().Handler(func(l net.Listener) error {
		var (
			conn net.Conn
			berr error
		)

		if conn, berr = l.Accept(); berr != nil {
			return berr
		}
		defer func() { _ = conn.Close() }()
		_, berr = bufio.NewReader(conn).ReadBytes(0)

		return berr
	})
	srv.Serve(tlsLst)
	defer srv.Stop()
	if _, err = New("tls://"+lst.Addr().String(), CompressionNone); err == nil {
		t.Errorf(tplFatalExpectingError)
	}
	// Ошибка рукопожатия TLS с сервером, немедленно закрывающим соединения.
	if lst, err = net.Listen(tcpNetwork, "127.0.0.1:0"); err != nil {
		t.Fatalf(tplFatalListenError, err)
	}
	defer func() { _ = lst.Close() }()
	go func() {
		var (
			conn   net.Conn
			accept error
		)

		for {
			if conn, accept = lst.Accept(); accept != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	if _, err = New("tls://"+lst.Addr().String(), CompressionNone); err == nil {
		t.Errorf(tplFatalExpectingError)
	}
}

// Функция установки TCP соединения с сервером по адресу.
func tcpDialFunc(address string) (ret func() (net.Conn, error)) {
	ret = func() (net.Conn, error) {
		var dialer net.Dialer

		return dialer.Dial(tcpNetwork, address)
	}

	return
}

// Приём данных одного TCP-соединения с ограничением времени ожидания.
func tcpReceive(lst net.Listener) (ret []byte, err error) {
	var (
		conn net.Conn
		buf  []byte
		n    int
	)

	if err = lst.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second)); err != nil {
		return
	}
	if conn, err = lst.Accept(); err != nil {
		return
	}
	defer func() { _ = conn.Close() }()
	buf = make([]byte, 4096)
	if n, err = conn.Read(buf); err != nil {
		return
	}
	ret = buf[:n]

	return
}

// Принятие и немедленное закрытие одного TCP-соединения.
func tcpDiscard(lst net.Listener) (err error) {
	var conn net.Conn

	if err = lst.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second)); err != nil {
		return
	}
	if conn, err = lst.Accept(); err != nil {
		return
	}
	err = conn.Close()

	return
}

// Создание самоподписанного сертификата и конфигураций TLS сервера и клиента для тестов.
func newTestTLSConfigs() (srv, cli *tls.Config, err error) {
	var (
		key  *ecdsa.PrivateKey
		tmpl *x509.Certificate
		der  []byte
		cert *x509.Certificate
		pool *x509.CertPool
	)

	if key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		return
	}
	tmpl = &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "gelf-test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}
	if der, err = x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key); err != nil {
		return
	}
	if cert, err = x509.ParseCertificate(der); err != nil {
		return
	}
	srv = &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
		MinVersion:   tlsMinVersion,
	}
	pool = x509.NewCertPool()
	pool.AddCert(cert)
	cli = &tls.Config{
		RootCAs:    pool,
		MinVersion: tlsMinVersion,
	}

	return
}
