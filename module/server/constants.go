package server

import "strings"

var (
	serverUnknown = serverType{0, ""}
	serverWeb     = serverType{1, "web"}
	serverGrpc    = serverType{2, "grpc"}
)

// Тип сервера.
type serverType struct {
	i uint16
	s string
}

// Преобразование uint16 в тип.
func serverTypeParseUint16(i uint16) (ret serverType) {
	switch i {
	case serverWeb.Uint16():
		ret = serverWeb
	case serverGrpc.Uint16():
		ret = serverGrpc
	default:
		ret = serverUnknown
	}

	return ret
}

// Преобразование строки в тип.
func serverTypeParseString(s string) (ret serverType) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case serverWeb.String():
		ret = serverWeb
	case serverGrpc.String():
		ret = serverGrpc
	default:
		ret = serverUnknown
	}

	return ret
}

// Uint16 Представление типа в качестве числа.
func (sto serverType) Uint16() uint16 { return sto.i }

// String Реализация интерфейса Stringify.
func (sto serverType) String() string { return sto.s }
