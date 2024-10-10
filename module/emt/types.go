package emt

import (
	"bytes"
	"io"
	"net/url"
	"regexp"

	"github.com/webnice/dic"
	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
)

const (
	size1gb      = 1024 * 1024 * 1024 * 1 // 1Гб памяти.
	keyExt       = "."                    // Разделитель для расширения имени файла.
	keyPathMacos = "__MACOSX"             // Техническое имя macOS.

)

var rexExt = regexp.MustCompile(`(?i)^.+\.(.+)?$`)

// Interface Интерфейс пакета.
type Interface interface {
	// ParseZip Загрузка и разбор шаблона из ZIP файла.
	ParseZip(file io.Reader) (ret *Template, err error)

	// Embedding Встраивание встраиваемого контента в шаблоны сообщений.
	Embedding(tpl *Template) (err error)

	// BodyImageUri Обход, найденных в теле шаблонов с типом HTML, ссылок на изображения и замена их.
	// Для каждого найденного изображения, происходит вызов функции, которая принимает решение о замене изображения.
	// В случае положительного решения, функции возвращает новый URI адрес изображения, который вставляется в шаблон.
	BodyImageUri(tpl *Template, fn ImageUriFn) (err error)

	// BodyImageEmbed Обход, найденных в теле шаблонов с типом HTML, встроенных изображений методом data:url, а так же
	// найденных в архиве встраиваемых изображений.
	// Для каждого изображения, происходит вызов функции, которая принимает решение о замене изображения.
	// В случае положительного решения, функции возвращает новый URI адрес изображения, который вставляется в шаблон.
	BodyImageEmbed(tpl *Template, fn ImageEmbedFn) (err error)
}

// Объект сущности, реализующий интерфейс пакета.
type impl struct {
	cfg kitModuleCfg.Interface // Конфигурация приложения.
}

// Описание замены в теле шаблона.
type replaceContent struct {
	Old  string    // Заменяемое значение.
	New  string    // Новое значение.
	Type dic.IMime // Тип контента.
}

// ImageUriFn Функция замены ссылок на изображение.
// Если функция вернёт пустое значение, тогда замена URI изображения, в теле шаблона, производиться не будет.
// Функция получает значения:
//
//	mme - MIME тип изображения;
//	now - Текущий URI картинки;
//
// Функция должна вернуть новое значение:
//
//	newUri - Новый URI изображения;
type ImageUriFn func(mme dic.IMime, now *url.URL) (newUri string)

// ImageEmbedFn Функция замены встроенных изображений методом data:url.
// Если функция вернёт пустое значение, тогда замена изображения, в теле шаблона, производиться не будет.
// Функция получает значения:
//
//	mme - MIME тип изображения;
//	img - Тело файла изображения;
//
// Функция должна вернуть новое значение:
//
//	newUri - Новый URI изображения;
type ImageEmbedFn func(mme dic.IMime, img *bytes.Buffer) (newUri string)
