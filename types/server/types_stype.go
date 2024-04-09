package server

import "strings"

var (
	TUnknown = Type{}
	TUdp     = Type{i: 1, s: "udp"}
	TTcp     = Type{i: 2, s: "tcp"}
	TWeb     = Type{i: 3, s: "web"}
	TGrpc    = Type{i: 4, s: "grpc"}
)

var _, _ = TypeParseUint16(0), TypeParseString("")

// Type Тип сервера.
type Type struct {
	i uint16
	s string
}

// TypeParseUint16 Преобразование uint16 в тип.
func TypeParseUint16(i uint16) (ret Type) {
	switch i {
	case TWeb.Uint16():
		ret = TWeb
	case TGrpc.Uint16():
		ret = TGrpc
	case TTcp.Uint16():
		ret = TTcp
	case TUdp.Uint16():
		ret = TUdp
	default:
		ret = TUnknown
	}

	return ret
}

// TypeParseString Преобразование строки в тип.
func TypeParseString(s string) (ret Type) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case TWeb.String():
		ret = TWeb
	case TGrpc.String():
		ret = TGrpc
	case TTcp.String():
		ret = TTcp
	case TUdp.String():
		ret = TUdp
	default:
		ret = TUnknown
	}

	return ret
}

// Uint16 Представление типа в качестве числа.
func (sto Type) Uint16() uint16 { return sto.i }

// String Реализация интерфейса Stringify.
func (sto Type) String() string { return sto.s }
