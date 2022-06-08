// Package tpl
package tpl

import (
	"bytes"
	"strings"
	"sync"

	kitTypes "github.com/webnice/kit/types"
)

// New Конструктор объекта сущности пакета, возвращается интерфейс пакета.
func New(out kitTypes.SyncWriter, s string) Interface {
	var tpl = &impl{
		src:   s,
		wr:    out,
		wrMux: new(sync.Mutex),
	}
	return tpl
}

// Parse Разбор шаблона.
func (tpl *impl) Parse() (err error) {
	var (
		ses    *session
		chunks []*chunk
		n      int
	)

	// Разрезание шаблона на куски.
	if chunks = tpl.splitChunk(tpl.src); len(chunks) == 0 {
		return
	}
	ses = &session{parent: tpl}
	tpl.checkChunk(chunks, ses) // Распознание кусков с типом chunkUnknown.
	for n = 1; n <= 4; n++ {
		switch n {
		case 1:
			tpl.formatCommentDelete(chunks) // Удаление всех тегов комментариев.
		case 2, 4:
			tpl.compactText(chunks) // Рядом стоящие куски с текстом объединяются в один кусок.
		case 3:
			tpl.formatChunk(chunks) // Применение тегов форматирования.
		}
		chunks = tpl.cleanDeleted(chunks) // Очистка удалённых кусков.
	}
	// Сохранение результата.
	tpl.chunks = make([]*chunk, 0, len(chunks))
	tpl.chunks = append(tpl.chunks, chunks...)
	// Создание объекта шаблонизатора.
	if tpl.tpl, err = tpl.Template(""); err != nil {
		return
	}

	return
}

// String Интерфейс Stringer. Представления шаблона в виде строки.
func (tpl *impl) String() (ret string) {
	var (
		n int
		s strings.Builder
	)

	for n = range tpl.chunks {
		_, _ = s.WriteString(tpl.chunks[n].String())
	}
	ret = s.String()

	return
}

// Output Вывод сообщения в интерфейс вывода.
func (tpl *impl) Output(buf *bytes.Buffer) {
	// Блокировка, так как в выходной поток может начать печатать другой процесс и в итоге на выходе будет каша.
	tpl.wrMux.Lock()
	_, _ = buf.WriteTo(tpl.wr)
	_, _ = tpl.wr.Write([]byte(tplNewLine))
	// Не забываем разблокировать, не используем defer для экономии микросекунд.
	tpl.wrMux.Unlock()
}
