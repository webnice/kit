package gelf

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestNewUdp(t *testing.T) {
	const (
		tplFatalUnexpectedError = "NewUdp() вернул неожиданную ошибку: %v."
		tplFatalExpectingError  = "NewUdp() должен был вернуть ошибку."
		tplErrorChunkSize       = "NewUdp() вернул размер чанка %d, ожидался %d."
	)
	var (
		err error
		cli ProtocolClient
		ucp *udpClient
	)

	// Размер чанка по умолчанию при нулевом значении размера чанка.
	if cli, err = NewUdp("127.0.0.1", 12201, 0, WithDialTimeout(time.Second)); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if ucp = cli.(*udpClient); ucp.chunkSize != DefaultChunkSize {
		t.Errorf(tplErrorChunkSize, ucp.chunkSize, DefaultChunkSize)
	}
	if err = cli.Close(); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	// Ошибка разбора адреса с недопустимым символом в имени хоста.
	if _, err = NewUdp("ho]st", 80, 0); err == nil {
		t.Errorf(tplFatalExpectingError)
	}
}

func TestUdpClientSendData(t *testing.T) {
	const (
		tplFatalListenError     = "net.ListenUDP() вернул неожиданную ошибку: %v."
		tplFatalUnexpectedError = "SendData() вернул неожиданную ошибку: %v."
		tplErrorChunkCount      = "получено %d чанков, ожидалось %d."
		tplErrorChunkMagic      = "чанк %d содержит неверный магический заголовок."
		tplErrorChunkSequence   = "чанк %d содержит неверную последовательность: индекс %d, количество %d."
		tplErrorMessageID       = "чанк %d содержит неверный идентификатор сообщения."
		tplErrorChunkData       = "чанк %d содержит неверные данные."
	)
	var (
		err        error
		lst        *net.UDPConn
		cli        ProtocolClient
		addr       *net.UDPAddr
		data       []byte
		chunk      []byte
		messageID  []byte
		chunkCount int
		n          int
	)

	if lst, err = net.ListenUDP(udpNetwork, &net.UDPAddr{IP: net.ParseIP("127.0.0.1")}); err != nil {
		t.Fatalf(tplFatalListenError, err)
	}
	defer func() { _ = lst.Close() }()
	addr = lst.LocalAddr().(*net.UDPAddr)
	if cli, err = NewUdp("127.0.0.1", uint16(addr.Port), 30); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	defer func() { _ = cli.Close() }()
	// Отправка сообщения в 100 байт чанками по 30 байт, ожидается 4 чанка.
	data = bytes.Repeat([]byte{0xAB}, 100)
	if err = cli.SendData(data); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	// Получение и проверка чанков.
	chunk = make([]byte, 1024)
	for n = 0; n < 4; n++ {
		var (
			length   int
			deadline time.Time
		)

		deadline = time.Now().Add(time.Second)
		if err = lst.SetReadDeadline(deadline); err != nil {
			t.Fatalf(tplFatalUnexpectedError, err)
		}
		if length, _, err = lst.ReadFromUDP(chunk); err != nil {
			t.Fatalf(tplFatalUnexpectedError, err)
		}
		if chunk[0] != chunkMagic0 || chunk[1] != chunkMagic1 {
			t.Errorf(tplErrorChunkMagic, n)
		}
		if n == 0 {
			messageID = chunk[2 : 2+messageIDSize]
		} else if !bytes.Equal(messageID, chunk[2:2+messageIDSize]) {
			t.Errorf(tplErrorMessageID, n)
		}
		if int(chunk[10]) != n || int(chunk[11]) != 4 {
			t.Errorf(tplErrorChunkSequence, n, chunk[10], chunk[11])
		}
		// Проверка полезных данных чанка.
		var begin, end int

		begin = n * 30
		end = begin + 30
		if end > len(data) {
			end = len(data)
		}
		if !bytes.Equal(data[begin:end], chunk[chunkHeaderSize:length]) {
			t.Errorf(tplErrorChunkData, n)
		}
		chunkCount++
	}
	if chunkCount != 4 {
		t.Errorf(tplErrorChunkCount, chunkCount, 4)
	}
}

func TestUdpClientSendDataEdge(t *testing.T) {
	const (
		tplFatalListenError     = "net.ListenUDP() вернул неожиданную ошибку: %v."
		tplFatalUnexpectedError = "SendData() вернул неожиданную ошибку: %v."
		tplErrorChunkSequence   = "чанк %d содержит неверную последовательность: индекс %d, количество %d."
		tplErrorChunkExtra      = "получен лишний чанк сообщения."
	)
	var (
		err      error
		lst      *net.UDPConn
		cli      ProtocolClient
		addr     *net.UDPAddr
		data     []byte
		chunk    []byte
		n        int
		deadline time.Time
	)

	if lst, err = net.ListenUDP(udpNetwork, &net.UDPAddr{IP: net.ParseIP("127.0.0.1")}); err != nil {
		t.Fatalf(tplFatalListenError, err)
	}
	defer func() { _ = lst.Close() }()
	addr = lst.LocalAddr().(*net.UDPAddr)
	if cli, err = NewUdp("127.0.0.1", uint16(addr.Port), 30); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	defer func() { _ = cli.Close() }()
	// Пустое сообщение не порождает отправку чанков.
	if err = cli.SendData(nil); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	// Размер сообщения кратен размеру чанка: 60 байт чанками по 30 байт, ожидается 2 чанка.
	data = bytes.Repeat([]byte{0xAB}, 60)
	if err = cli.SendData(data); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	chunk = make([]byte, 1024)
	for n = 0; n < 2; n++ {
		deadline = time.Now().Add(time.Second)
		if err = lst.SetReadDeadline(deadline); err != nil {
			t.Fatalf(tplFatalUnexpectedError, err)
		}
		if _, _, err = lst.ReadFromUDP(chunk); err != nil {
			t.Fatalf(tplFatalUnexpectedError, err)
		}
		if int(chunk[10]) != n || int(chunk[11]) != 2 {
			t.Errorf(tplErrorChunkSequence, n, chunk[10], chunk[11])
		}
	}
	// Отсутствие дополнительных чанков.
	deadline = time.Now().Add(50 * time.Millisecond)
	if err = lst.SetReadDeadline(deadline); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if _, _, err = lst.ReadFromUDP(chunk); err == nil {
		t.Errorf(tplErrorChunkExtra)
	}
}

func TestUdpClientSendDataTooLarge(t *testing.T) {
	const (
		tplFatalUnexpectedError = "NewUdp() вернул неожиданную ошибку: %v."
		tplFatalExpectingError  = "SendData() должен был вернуть ошибку."
	)
	var (
		err  error
		cli  ProtocolClient
		data []byte
	)

	if cli, err = NewUdp("127.0.0.1", 12201, 10); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	defer func() { _ = cli.Close() }()
	// Сообщение в 1281 байт превышает максимально допустимые 128 чанков по 10 байт.
	data = make([]byte, 1281)
	if err = cli.SendData(data); err == nil {
		t.Errorf(tplFatalExpectingError)
	} else if !Errors().MessageTooLarge.Is(err) {
		t.Errorf(tplFatalUnexpectedError, err)
	}
}

func TestUdpClientSendDataError(t *testing.T) {
	const (
		tplFatalUnexpectedError = "NewUdp() вернул неожиданную ошибку: %v."
		tplFatalExpectingError  = "SendData() должен был вернуть ошибку."
	)
	var (
		err error
		cli ProtocolClient
	)

	// Ошибка отправки данных по закрытому соединению.
	if cli, err = NewUdp("127.0.0.1", 12201, 0); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	_ = cli.Close()
	if err = cli.SendData([]byte("data")); err == nil {
		t.Errorf(tplFatalExpectingError)
	}
	// Ошибка отправки датаграммы, превышающей максимальный размер UDP-сообщения.
	if cli, err = NewUdp("127.0.0.1", 9, 70000); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if err = cli.SendData(make([]byte, 70000)); err == nil {
		t.Errorf(tplFatalExpectingError)
	}
	_ = cli.Close()
}
