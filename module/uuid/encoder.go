package uuid

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// FromBytes Конвертация среза байт в UUID, в случае ошибки возвращается константа NULL.
func (ui *impl) FromBytes(input []byte) (ret UUID) {
	var (
		u   *uuid
		err error
	)

	u = &uuid{}
	if err = u.UnmarshalBinary(input); err != nil {
		ret = NULL
		return
	}
	ret = u

	return
}

// FromString Конвертация строки в UUID, в случае ошибки возвращается константа NULL.
func (ui *impl) FromString(input string) (ret UUID) {
	var (
		u   *uuid
		err error
	)

	u = &uuid{}
	if err = u.UnmarshalText([]byte(input)); err != nil {
		ret = NULL
		return
	}
	ret = u

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
func (u *uuid) MarshalBinary() (data []byte, err error) { data = u.Bytes(); return }

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (u *uuid) UnmarshalBinary(data []byte) (err error) {
	const lengthError = "UUID должен иметь длину 16 байт"

	if len(data) != size {
		err = errors.New(lengthError)
		return
	}
	copy(u.data[:], data)

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
func (u *uuid) MarshalText() (text []byte, err error) { text = []byte(u.String()); return }

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
// Поддерживаются форматы:
//
//	"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
//	"{6ba7b810-9dad-11d1-80b4-00c04fd430c8}",
//	"urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8",
//	"6ba7b8109dad11d180b400c04fd430c8".
//
//	uuid := canonical | hashlike | braced | urn
//	plain := canonical | hashlike
//	canonical := 4hexoct '-' 2hexoct '-' 2hexoct '-' 6hexoct
//	hashlike := 12hexoct
//	braced := '{' plain '}'
//	urn := URN ':' UUID-NID ':' plain
//	URN := 'urn'
//	UUID-NID := 'uuid'
//	12hexoct := 6hexoct 6hexoct
//	6hexoct := 4hexoct 2hexoct
//	4hexoct := 2hexoct 2hexoct
//	2hexoct := hexoct hexoct
//	hexoct := hexdig hexdig
//	hexdig := '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' |
//	          'a' | 'b' | 'c' | 'd' | 'e' | 'f' |
//	          'A' | 'B' | 'C' | 'D' | 'E' | 'F'
func (u *uuid) UnmarshalText(text []byte) (err error) {
	switch len(text) {
	case 32:
		return u.decodeHashLike(text)
	case 34:
		fallthrough
	case 36:
		if len(text) == 34 {
			return u.decodeBraced(text)
		}
		return u.decodeCanonical(text)
	case 38:
		return u.decodeBraced(text)
	case 41:
		fallthrough
	case 45:
		return u.decodeURN(text)
	default:
		return fmt.Errorf("uuid: некорректная длина UUID: %s", text)
	}
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (u *uuid) UnmarshalJSON(b []byte) (err error) {
	var text string

	if err = json.Unmarshal(b, &text); err != nil {
		return
	}
	err = u.UnmarshalText([]byte(text))

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
func (u *uuid) MarshalJSON() (ret []byte, err error) { return json.Marshal(u.String()) }

// decodeCanonical декодирует UUID из формата "6ba7b810-9dad-11d1-80b4-00c04fd430c8".
func (u *uuid) decodeCanonical(t []byte) (err error) {
	const formatError = "неверный формат UUID"
	var (
		src       []byte
		dst       []byte
		byteGroup int
		n         int
	)

	if t[8] != '-' || t[13] != '-' || t[18] != '-' || t[23] != '-' {
		err = errors.New(formatError)
		return
	}
	src, dst = t[:], u.data[:]
	for n, byteGroup = range byteGroups {
		if n > 0 {
			src = src[1:]
		}
		if _, err = hex.Decode(dst[:byteGroup/2], src[:byteGroup]); err != nil {
			return
		}
		src, dst = src[byteGroup:], dst[byteGroup/2:]
	}

	return
}

// decodeHashLike декодирует UUID из формата "6ba7b8109dad11d180b400c04fd430c8".
func (u *uuid) decodeHashLike(t []byte) (err error) {
	var (
		src []byte
		dst []byte
	)

	src, dst = t[:], u.data[:]
	if _, err = hex.Decode(dst, src); err != nil {
		return
	}

	return
}

// decodeBraced декодирует UUID из формата "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}" или "{6ba7b8109dad11d180b400c04fd430c8}".
func (u *uuid) decodeBraced(t []byte) (err error) {
	const formatError = "неверный формат UUID"
	var (
		l int
	)

	if l = len(t); t[0] != '{' || t[l-1] != '}' {
		err = errors.New(formatError)
		return
	}
	err = u.decodePlain(t[1 : l-1])

	return
}

// decodePlain декодирует UUID из формата "6ba7b810-9dad-11d1-80b4-00c04fd430c8" или "6ba7b8109dad11d180b400c04fd430c8".
func (u *uuid) decodePlain(t []byte) (err error) {
	const formatError = "неверный формат UUID"

	switch len(t) {
	case 32:
		err = u.decodeHashLike(t)
	case 36:
		err = u.decodeCanonical(t)
	default:
		err = errors.New(formatError)
	}

	return
}

// decodeURN декодирует UUID из формата "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8" или
// "urn:uuid:6ba7b8109dad11d180b400c04fd430c8".
func (u *uuid) decodeURN(t []byte) (err error) {
	const formatError = "неверный формат UUID"
	var (
		total         int
		urnUUIDPrefix []byte
	)

	total, urnUUIDPrefix = len(t), t[:9]
	if !bytes.Equal(urnUUIDPrefix, urnPrefix) {
		err = errors.New(formatError)
		return
	}
	err = u.decodePlain(t[9:total])

	return
}
