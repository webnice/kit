package tpl

import (
	"strconv"
	"strings"
)

// Разрезание шаблона на куски.
func (tpl *impl) splitChunk(s string) (ret []*chunk) {
	var (
		chk  *chunk
		idx  [][]int
		p, n int
	)

	idx = rexTag.FindAllStringSubmatchIndex(s, -1)
	for n = range idx {
		// Добавление куска с текстом.
		if idx[n][0] != p {
			ret = append(ret, &chunk{Type: chunkText, Src: s[p:idx[n][0]]})
			p = idx[n][1]
		}
		// Добавление куска с тегом.
		chk = &chunk{
			Src:  strings.ToLower(s[idx[n][2]:idx[n][3]]),
			Type: chunkUnknown,
		}
		if idx[n][4] != -1 && idx[n][5] != -1 {
			chk.Arg = s[idx[n][4]:idx[n][5]]
		}
		ret, p = append(ret, chk), idx[n][1]
	}
	// Добавление куска с текстом после последнего куска распознанного регулярным выражением.
	if len(s) > p {
		ret = append(ret, &chunk{Type: chunkText, Src: s[p:]})
	}

	return
}

// Распознание кусков с типом chunkUnknown.
// Куски с неизвестными тегами, конвертируются в куски с исходным текстом.
func (tpl *impl) checkChunk(chunks []*chunk, ses *session) {
	var (
		tagData   map[string]*tagDataInfo
		tagColor  map[string]*tagDataInfo
		tagFormat map[string]*tagFormatInfo
		n         int
		key       string
		found     bool
	)

	tagData, tagColor, tagFormat = ses.tagData(), ses.tagColor(), ses.tagFormat()
	for n = range chunks {
		if chunks[n].Type != chunkUnknown {
			continue
		}
		found = false
		// Теги данных.
		for key = range tagData {
			if chunks[n].Src == key {
				chunks[n].Type, chunks[n].Tag, found = chunkData, tagData[key], true
				break
			}
		}
		if found {
			continue
		}
		// Теги переключения цвета.
		for key = range tagColor {
			if chunks[n].Src == key {
				chunks[n].Type, chunks[n].Tag, found = chunkColor, tagColor[key], true
				break
			}
		}
		if found {
			continue
		}
		// Теги форматирования.
		for key = range tagFormat {
			if chunks[n].Src == key {
				chunks[n].Type, chunks[n].Tag, found = chunkFormat, tagFormat[key], true
				break
			}
		}
		if found {
			continue
		}
		// Куски с неизвестными тегами, конвертируются в куски с исходным текстом.
		chunks[n].Src = chunks[n].String()
		chunks[n].Arg, chunks[n].Type = "", chunkText
	}

	return
}

// Рядом стоящие куски с текстом объединяются в один кусок.
func (tpl *impl) compactText(chunks []*chunk) {
	var n, p int

	for n = range chunks {
		if chunks[n].Type != chunkText {
			continue
		}
		if len(chunks) <= n+1 {
			continue
		}
		for p = n + 1; p < len(chunks); p++ {
			if chunks[p].Type != chunkText {
				break
			}
			chunks[n].Src, chunks[p].Type = chunks[n].Src+chunks[p].Src, chunkDelete
		}
	}
}

// Очистка удалённых кусков.
func (tpl *impl) cleanDeleted(chunks []*chunk) (ret []*chunk) {
	var (
		n     int
		count int
	)

	for n = range chunks {
		if chunks[n].Type != chunkDelete {
			count++
		}
	}
	ret = make([]*chunk, 0, count)
	for n = range chunks {
		if chunks[n].Type == chunkDelete {
			continue
		}
		ret = append(ret, chunks[n])
	}

	return
}

// Удаление всех тегов комментариев.
func (tpl *impl) formatCommentDelete(chunks []*chunk) (ret []*chunk) {
	var (
		fi *tagFormatInfo
		n  int
	)

	for n = range chunks {
		if chunks[n].Type != chunkFormat {
			continue
		}
		switch ti := chunks[n].Tag.(type) {
		case *tagFormatInfo:
			fi = ti
		default:
			continue
		}
		switch fi.Name {
		case "#":
			chunks[n].Src, chunks[n].Type = "", chunkDelete
		}
	}

	return
}

// Применение тегов форматирования.
func (tpl *impl) formatChunk(chunks []*chunk) {
	var (
		err   error
		count uint64
		n, p  int
		fi    *tagFormatInfo
	)

	for n = range chunks {
		if chunks[n].Type != chunkFormat {
			continue
		}
		switch ti := chunks[n].Tag.(type) {
		case *tagFormatInfo:
			fi = ti
		default:
			continue
		}
		switch fi.Name {
		case "#":
			chunks[n].Src, chunks[n].Type = "", chunkDelete
		case "bp--":
			// ${bp--} - Удаление тега и переносов строки
			if n+1 < len(chunks) {
				for p = n + 1; p < len(chunks); p++ {
					if chunks[p].Type == chunkDelete {
						continue
					}
					if chunks[p].Type != chunkText {
						break
					}
					if chunks[p].Src = rexBrFirst.ReplaceAllString(chunks[p].Src, ""); chunks[p].Src != "" {
						break
					}
					chunks[p].Type = chunkDelete
				}
			}
			chunks[n].Type = chunkDelete
		case "bp-+":
			// ${bp-+} - Вставка переноса строки
			if count, err = strconv.ParseUint(chunks[n].Arg, 10, 16); err != nil || count == 0 {
				count = 1
			}
			chunks[n].Src, chunks[n].Type = strings.Repeat("\n", int(count)), chunkText
		case "spc--":
			// ${spc--} - Удаление тега и всех пробельных символов идущих после тега.
			if count, err = strconv.ParseUint(chunks[n].Arg, 10, 16); err != nil {
				count = 0
			}
			if n+1 < len(chunks) {
				for p = n + 1; p < len(chunks); p++ {
					if chunks[p].Type == chunkDelete {
						continue
					}
					if chunks[p].Type != chunkText {
						break
					}
					if chunks[p].Src = rexSpaceFirst.ReplaceAllString(chunks[p].Src, ""); chunks[p].Src != "" {
						break
					}
					chunks[p].Type = chunkDelete
				}
			}
			if chunks[n].Type = chunkDelete; count > 0 {
				chunks[n].Type, chunks[n].Arg, chunks[n].Src = chunkText, "", strings.Repeat(" ", int(count))
			}
		case "--spc":
			// ${--spc} - Удаление тега и всех пробельных символов идущих перед тегом.
			if count, err = strconv.ParseUint(chunks[n].Arg, 10, 16); err != nil {
				count = 0
			}
			if n > 0 {
				for p = n - 1; p >= 0; p-- {
					if chunks[p].Type == chunkDelete {
						continue
					}
					if chunks[p].Type != chunkText {
						break
					}
					if chunks[p].Src = rexSpaceLast.ReplaceAllString(chunks[p].Src, ""); chunks[p].Src != "" {
						break
					}
					chunks[p].Type = chunkDelete
				}
			}
			if chunks[n].Type = chunkDelete; count > 0 {
				chunks[n].Type, chunks[n].Arg, chunks[n].Src = chunkText, "", strings.Repeat(" ", int(count))
			}
		case "-spc-":
			// ${-spc-} - Удаление тега и всех пробельных символов идущих как перед тегом, так и после тега.
			if count, err = strconv.ParseUint(chunks[n].Arg, 10, 16); err != nil {
				count = 0
			}
			if n > 0 {
				for p = n - 1; p >= 0; p-- {
					if chunks[p].Type == chunkDelete {
						continue
					}
					if chunks[p].Type != chunkText {
						break
					}
					if chunks[p].Src = rexSpaceLast.ReplaceAllString(chunks[p].Src, ""); chunks[p].Src != "" {
						break
					}
					chunks[p].Type = chunkDelete
				}
			}
			if n+1 < len(chunks) {
				for p = n + 1; p < len(chunks); p++ {
					if chunks[p].Type == chunkDelete {
						continue
					}
					if chunks[p].Type != chunkText {
						break
					}
					if chunks[p].Src = rexSpaceFirst.ReplaceAllString(chunks[p].Src, ""); chunks[p].Src != "" {
						break
					}
					chunks[p].Type = chunkDelete
				}
			}
			if chunks[n].Type = chunkDelete; count > 0 {
				chunks[n].Type, chunks[n].Arg, chunks[n].Src = chunkText, "", strings.Repeat(" ", int(count))
			}
		}
	}
}
