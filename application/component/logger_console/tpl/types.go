// Package tpl
package tpl

import (
	"bytes"
	"io"
	"regexp"
	"sync"
	"text/template"

	kitModuleLog "github.com/webnice/kit/module/log"

	"github.com/muesli/termenv"
)

// Регулярное выражение для извлечения из шаблона всех тегов.
var (
	rexTag             = regexp.MustCompile(`(?mi)\${([a-z0-9-+#]+)(?::(.*?[^\\]))?}`)
	rexBrFirst         = regexp.MustCompile(`(?i)^([\r\n]+)`)
	rexSpaceFirst      = regexp.MustCompile(`(?i)^([[:space:]]+)`)
	rexSpaceLast       = regexp.MustCompile(`(?i)([[:space:]]+)$`)
	rexCurlyBraceOpen  = regexp.MustCompile(`(?i)({+)`)
	rexCurlyBraceClose = regexp.MustCompile(`(?i)(}+)`)
	rexHexColor        = regexp.MustCompile(`(?im)^(#[0-9a-f]{6})$`)
)

// Interface Интерфейс пакета.
type Interface interface {
	// Parse Разбор шаблона.
	Parse() (err error)

	// String Интерфейс Stringer. Представления шаблона в виде строки.
	String() (ret string)

	// NewSession Создание сесии обработки данных по шаблону.
	NewSession(data *kitModuleLog.Message) (ret Session, err error)

	// Output Вывод сообщения в интерфейс вывода.
	Output(buf *bytes.Buffer)
}

// Session Сессия обработки данных.
type Session interface {
	// Do Выполнение обработки данных через шаблонизатор.
	Do() (ret *bytes.Buffer, err error)
}

// Объект сущности, реализующий интерфейс Interface.
type impl struct {
	src    string             // Исходный шаблон.
	chunks []*chunk           // Куски исходного шаблона.
	tpl    *template.Template // Шаблон сообщения в виде объекта шаблонизатора.
	wr     io.Writer          // Интерфейс вывода потоковых сообщений.
	wrMux  *sync.Mutex        // Обеспечение атомарности вывода потоковых сообщений.
}

// Объект сессии.
type session struct {
	parent  *impl                 // Ссылка на родительский объект.
	writer  *bytes.Buffer         // Буфер писателя.
	profile termenv.Profile       // Профайл терминала, конвертация цветов.
	Data    *kitModuleLog.Message // Исходные данные для обработки.
	Tpl     *template.Template    // Копия шаблона для обработки данных в пределах сессии.
}

// Информация о тегах данных и цвета.
type tagDataInfo struct {
	Name    string      // Название функции шаблонизатора.
	Func    interface{} // Функция - обработчик.
	Docs    []string    // Документация по тегу.
	Example string      // Пример указания тега.
}

// Информация о тегах форматирования.
type tagFormatInfo struct {
	Name    string   // Название тега.
	Docs    []string // Документация по тегу.
	Example string   // Пример указания тега.
}
