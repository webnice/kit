package resource

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// Тип контента тестовых ресурсов.
const testResourceContentType = "application/octet-stream"

func testResource(content []byte) Resource {
	return Resource{
		Size:        uint64(len(content)),
		Time:        time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC),
		ContentType: testResourceContentType,
		Content:     content,
	}
}

func readTarFiles(t *testing.T, reader io.Reader) (content map[string][]byte, modes map[string]int64) {
	const (
		tplTarNextError    = "tar.Next() вернул ошибку: %v."
		tplTarReadAllError = "tar.ReadAll() вернул ошибку: %v."
	)
	var (
		tr  *tar.Reader
		hdr *tar.Header
		buf []byte
		err error
	)

	tr = tar.NewReader(reader)
	content, modes = make(map[string][]byte), make(map[string]int64)
	for {
		if hdr, err = tr.Next(); errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf(tplTarNextError, err)
		}
		if buf, err = io.ReadAll(tr); err != nil {
			t.Fatalf(tplTarReadAllError, err)
		}
		content[hdr.Name] = buf
		modes[hdr.Name] = hdr.Mode
	}

	return
}

func TestImplementationConstructor(t *testing.T) {
	const (
		errConstructorNil       = "Constructor() вернул nil."
		errConstructorNoMap     = "Constructor() не создал карту ресурсов."
		errConstructorHasGroups = "новый объект не должен содержать групп ресурсов."
	)
	var impl *Implementation

	impl = Constructor()
	if impl == nil {
		t.Fatal(errConstructorNil)
	}
	if impl.Res == nil {
		t.Fatal(errConstructorNoMap)
	}
	if len(impl.Group()) != 0 {
		t.Fatal(errConstructorHasGroups)
	}
}

func TestImplementationAddAndResourceData(t *testing.T) {
	const (
		tplAddError                  = "Add() вернул ошибку: %v."
		errResourceDataNil           = "ResourceData() вернул nil после Add()."
		errResourceDataNotCopy       = "ResourceData() вернул исходный объект ресурса, ожидалась копия."
		tplResourceDataUnexpected    = "ResourceData() вернул описание %+v, ожидалось %+v."
		tplContentUnexpected         = "ResourceData().Content = %v, ожидалось %v."
		errAddNotCopied              = "Add() не скопировал контент ресурса."
		errResourceDataNoCopy        = "ResourceData() вернул контент без копирования."
		tplGroupUnexpected           = "Group() = %v, ожидалась группа assets."
		tplResourceByGroupUnexpected = "ResourceByGroup() = %v, ожидался ресурс bin/data.bin."
		errUnknownGroupNotNil        = "ResourceData() для неизвестной группы вернул не nil."
		errUnknownNameNotNil         = "ResourceData() для неизвестного ресурса вернул не nil."
	)
	var (
		impl    *Implementation
		content []byte
		res     Resource
		got     *Resource
		stored  *Resource
		groups  []string
		names   []string
		err     error
	)

	impl = Constructor()
	content = []byte{0x01, 0x02, 0x03, 0x00}
	res = testResource(content)
	if err = impl.Add("assets", "bin/data.bin", res); err != nil {
		t.Fatalf(tplAddError, err)
	}
	got = impl.ResourceData("assets", "bin/data.bin")
	if got == nil {
		t.Fatal(errResourceDataNil)
	}
	if got == &res {
		t.Fatal(errResourceDataNotCopy)
	}
	if got.Size != res.Size || !got.Time.Equal(res.Time) || got.ContentType != res.ContentType {
		t.Fatalf(tplResourceDataUnexpected, got, res)
	}
	if !bytes.Equal(got.Content, content) {
		t.Fatalf(tplContentUnexpected, got.Content, content)
	}
	res.Content[0] = 0xff
	stored = impl.ResourceData("assets", "bin/data.bin")
	if bytes.Equal(stored.Content, res.Content) {
		t.Fatal(errAddNotCopied)
	}
	got.Content[1] = 0xff
	stored = impl.ResourceData("assets", "bin/data.bin")
	if stored.Content[1] == 0xff {
		t.Fatal(errResourceDataNoCopy)
	}
	groups = impl.Group()
	if len(groups) != 1 || groups[0] != "assets" {
		t.Fatalf(tplGroupUnexpected, groups)
	}
	names = impl.ResourceByGroup("assets")
	if len(names) != 1 || names[0] != "bin/data.bin" {
		t.Fatalf(tplResourceByGroupUnexpected, names)
	}
	if impl.ResourceData("unknown", "bin/data.bin") != nil {
		t.Fatal(errUnknownGroupNotNil)
	}
	if impl.ResourceData("assets", "unknown") != nil {
		t.Fatal(errUnknownNameNotNil)
	}
}

func TestImplementationAddErrors(t *testing.T) {
	const (
		tplAddError           = "Add() вернул ошибку: %v."
		errAddDuplicateNil    = "Add() при дублировании ресурса вернул nil ошибку."
		errAddDuplicateChange = "дублирующий Add() изменил существующий ресурс."
		errAddBadSizeNil      = "Add() с неверным размером вернул nil ошибку."
		tplAddBadSizeLeft     = "Add() с неверным размером оставил ресурс: %+v."
		tplAddBadSizeGroup    = "Add() с неверным размером создал группу с ресурсами: %v."
		errAddBadSizeNilEx    = "Add() с неверным размером в существующей группе вернул nil ошибку."
		tplAddBadSizeLeftEx   = "Add() с неверным размером в существующей группе оставил ресурс: %+v."
	)
	var (
		impl           *Implementation
		res            Resource
		stored         *Resource
		brokenNewGroup Resource
		broken         Resource
		names          []string
		got            *Resource
		err            error
	)

	impl = Constructor()
	res = testResource([]byte("valid"))
	if err = impl.Add("assets", "file.txt", res); err != nil {
		t.Fatalf(tplAddError, err)
	}
	if err = impl.Add("assets", "file.txt", res); err == nil {
		t.Fatal(errAddDuplicateNil)
	}
	//goland:noinspection GoMaybeNil
	stored = impl.ResourceData("assets", "file.txt")
	if !bytes.Equal(stored.Content, res.Content) {
		t.Fatal(errAddDuplicateChange)
	}
	brokenNewGroup = testResource([]byte("123"))
	brokenNewGroup.Size = 10
	if err = impl.Add("broken", "bad-size", brokenNewGroup); err == nil {
		t.Fatal(errAddBadSizeNil)
	}
	got = impl.ResourceData("broken", "bad-size")
	if got != nil {
		t.Fatalf(tplAddBadSizeLeft, got)
	}
	names = impl.ResourceByGroup("broken")
	if len(names) != 0 {
		t.Fatalf(tplAddBadSizeGroup, names)
	}
	broken = testResource([]byte("123"))
	broken.Size = 10
	if err = impl.Add("assets", "bad-size", broken); err == nil {
		t.Fatal(errAddBadSizeNilEx)
	}
	got = impl.ResourceData("assets", "bad-size")
	if got != nil {
		t.Fatalf(tplAddBadSizeLeftEx, got)
	}
}

func TestImplementationNilResourceData(t *testing.T) {
	const (
		tplNilResourceNotNil   = "ResourceData() для nil ресурса вернул %+v."
		tplNilResourceError    = "ResourceByGroupTarReader() для nil ресурса вернул ошибку: %v."
		tplNilResourceHasFiles = "ResourceByGroupTarReader() для nil ресурса вернул файлы: %v."
	)
	var (
		impl    *Implementation
		reader  *bytes.Reader
		content map[string][]byte
		got     *Resource
		err     error
	)

	impl = Constructor()
	impl.Res["nil"] = map[string]*Resource{"item": nil}
	got = impl.ResourceData("nil", "item")
	if got != nil {
		t.Fatalf(tplNilResourceNotNil, got)
	}
	reader, err = impl.ResourceByGroupTarReader("nil")
	if err != nil {
		t.Fatalf(tplNilResourceError, err)
	}
	content, _ = readTarFiles(t, reader)
	if len(content) != 0 {
		t.Fatalf(tplNilResourceHasFiles, content)
	}
}

func TestImplementationResourceByGroupTarReader(t *testing.T) {
	const (
		tplAddError            = "Add(%q) вернул ошибку: %v."
		tplTarReaderError      = "ResourceByGroupTarReader() вернул ошибку: %v."
		errTarReaderNil        = "ResourceByGroupTarReader() вернул nil reader."
		tplTarFilesCount       = "tar контейнер содержит %d файла, ожидалось 3: %v."
		tplTarFilesContent     = "tar контейнер содержит неверный контент: %v."
		tplTarFileMode         = "mode файла %q = %o, ожидался 0600."
		tplTarReaderUnknownErr = "ResourceByGroupTarReader() для неизвестной группы вернул ошибку: %v."
		tplTarUnknownHasFiles  = "tar контейнер для неизвестной группы содержит файлы: %v."
	)
	var (
		impl     *Implementation
		text     []byte
		bin      []byte
		empty    []byte
		emptyRes Resource
		reader   *bytes.Reader
		content  map[string][]byte
		modes    map[string]int64
		item     struct {
			name string
			res  Resource
		}
		name string
		mode int64
		err  error
	)

	impl = Constructor()
	text = []byte("hello")
	bin = []byte{0x00, 0xff, 0x10}
	empty = nil
	emptyRes = testResource(empty)
	emptyRes.ContentType = "text/plain"
	for _, item = range []struct {
		name string
		res  Resource
	}{
		{name: "a.txt", res: testResource(text)},
		{name: "dir/b.bin", res: testResource(bin)},
		{name: "empty.txt", res: emptyRes},
	} {
		if err = impl.Add("assets", item.name, item.res); err != nil {
			t.Fatalf(tplAddError, item.name, err)
		}
	}
	reader, err = impl.ResourceByGroupTarReader("assets")
	if err != nil {
		t.Fatalf(tplTarReaderError, err)
	}
	if reader == nil {
		t.Fatal(errTarReaderNil)
	}
	content, modes = readTarFiles(t, reader)
	if len(content) != 3 {
		t.Fatalf(tplTarFilesCount, len(content), content)
	}
	if !bytes.Equal(content["a.txt"], text) || !bytes.Equal(content["dir/b.bin"], bin) || !bytes.Equal(content["empty.txt"], empty) {
		t.Fatalf(tplTarFilesContent, content)
	}
	for name, mode = range modes {
		if mode != 0600 {
			t.Fatalf(tplTarFileMode, name, mode)
		}
	}
	reader, err = impl.ResourceByGroupTarReader("unknown")
	if err != nil {
		t.Fatalf(tplTarReaderUnknownErr, err)
	}
	content, _ = readTarFiles(t, reader)
	if len(content) != 0 {
		t.Fatalf(tplTarUnknownHasFiles, content)
	}
}

func TestImplementationResourceByGroupTarReaderErrors(t *testing.T) {
	const (
		errTarReaderNilError   = "ResourceByGroupTarReader() вернул nil ошибку."
		errTarReaderNotNil     = "ResourceByGroupTarReader() при ошибке вернул не nil reader."
		tplTarReaderWrongError = "ResourceByGroupTarReader() вернул ошибку %q, ожидалось %q."
	)
	var tests = map[string]struct {
		size    uint64
		content []byte
		wantErr string
	}{
		"header": {
			size:    uint64(1) << 63,
			content: []byte{0x01},
			wantErr: "запись заголовка файла",
		},
		"body": {
			size:    1,
			content: []byte{0x01, 0x02, 0x03},
			wantErr: "запись тела файла",
		},
		"close": {
			size:    10,
			content: []byte{0x01},
			wantErr: "в момент закрытия",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				impl   *Implementation
				reader *bytes.Reader
				err    error
			)

			impl = Constructor()
			impl.Res["assets"] = map[string]*Resource{
				"broken": {
					Size:    test.size,
					Time:    time.Now(),
					Content: test.content,
				},
			}
			reader, err = impl.ResourceByGroupTarReader("assets")
			if err == nil {
				t.Fatal(errTarReaderNilError)
			}
			if reader != nil {
				t.Fatal(errTarReaderNotNil)
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf(tplTarReaderWrongError, err.Error(), test.wantErr)
			}
		})
	}
}

func TestImplementationConcurrentAccess(t *testing.T) {
	const (
		tplNameTemplate    = "dir/%d/%d.txt"
		tplConcurrentError = "concurrent access вернул ошибку: %v."
	)
	var (
		err   error
		n, m  int
		impl  *Implementation
		wg    sync.WaitGroup
		errCh chan error
	)

	impl = Constructor()
	wg = sync.WaitGroup{}
	errCh = make(chan error, 128)
	for n = range 8 {
		wg.Add(1)
		go func(n int) {
			const (
				tplAddError       = "Add() вернул ошибку: %v."
				tplTarReaderError = "ResourceByGroupTarReader() вернул ошибку: %v."
			)
			var (
				name    string
				content []byte
				err     error
			)

			defer wg.Done()
			for m = range 20 {
				name = fmt.Sprintf(tplNameTemplate, n, m)
				content = []byte(name)
				if err = impl.Add("concurrent", name, testResource(content)); err != nil {
					errCh <- err

					return
				}
				impl.Group()
				impl.ResourceByGroup("concurrent")
				impl.ResourceData("concurrent", name)
				if _, err = impl.ResourceByGroupTarReader("concurrent"); err != nil {
					errCh <- err

					return
				}
			}
		}(n)
	}
	wg.Wait()
	close(errCh)
	for err = range errCh {
		t.Fatalf(tplConcurrentError, err)
	}
}
