package emt

import (
	"encoding/json"
	"net/url"
	"path"
	"strings"
)

// Template Структура объекта электронного сообщения.
type Template struct {
	Body       []Body     `json:"body"`     // Массив шаблонов электронного сообщения.
	Tags       []Tags     `json:"tags"`     // Массив тегов электронного сообщения.
	Embedded   []File     `json:"embedded"` // Массив встраиваемых вложений электронного сообщения.
	Attach     []File     `json:"attach"`   // Массив прикрепляемых вложений электронного сообщения.
	BodyImgUri []*url.URL `json:"-"`        // Найденные в теле шаблона ссылки на изображения.
}

// Body Часть электронного сообщения - шаблоны.
type Body struct {
	Type  string `json:"type"`  // Тип контента.
	Value string `json:"value"` // Тело письма в кодировке BASE64.
}

// Tags Часть электронного сообщения - теги.
type Tags struct {
	Key   string `json:"key"`   // Название тега, латиница, с заглавной буквы.
	Value string `json:"value"` // Значение тега по умолчанию.
}

// File Часть электронного сообщения - вложения.
type File struct {
	Type  string `json:"type"`  // Тип контента.
	Name  string `json:"name"`  // Название объекта.
	Value string `json:"value"` // Тело объекта в кодировке BASE64.
}

// ValueGet Декодирование значения Value из BASE64 в срез байт.
func (tbo *Body) ValueGet() (ret []byte) { return decodeBase64(tbo.Value) }

// ValueSet Кодирование среда байт в значение BASE64 и установка в свойство Value объекта.
func (tbo *Body) ValueSet(b []byte) { tbo.Value = encodeBase64(b) }

// ValueGet Декодирование значения Value из BASE64 в срез байт.
func (tfo *File) ValueGet() (ret []byte) { return decodeBase64(tfo.Value) }

// ValueSet Кодирование среда байт в значение BASE64 и установка в свойство Value объекта.
func (tfo *File) ValueSet(b []byte) { tfo.Value = encodeBase64(b) }

// JsonUnmarshal Декодирование значения Value из BASE64 в срез байт, а затем декодирование среза
// байт из JSON в переданный объект.
func (tfo *File) JsonUnmarshal(o any) error { return json.Unmarshal(tfo.ValueGet(), o) }

// JsonMarshal Кодирование переданного объекта в JSON и запись полученного среза байт в
// кодировке BASE64 в значение поля value.
func (tfo *File) JsonMarshal(o any) (err error) {
	var buf []byte

	if buf, err = json.Marshal(o); err != nil {
		return
	}
	tfo.ValueSet(buf)

	return
}

// Filename Выделение из названия объекта, наименование файла.
func (tfo *File) Filename() (ret string) {
	_, ret = path.Split(tfo.Name)
	return
}

// FilenameExtension Выделение из названия объекта расширения наименование файла.
func (tfo *File) FilenameExtension() string {
	return strings.TrimPrefix(path.Ext(tfo.Filename()), ".")
}
