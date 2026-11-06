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

func testResource(content []byte) Resource {
	return Resource{
		Size:        uint64(len(content)),
		Time:        time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC),
		ContentType: "application/octet-stream",
		Content:     content,
	}
}

func readTarFiles(t *testing.T, reader io.Reader) (content map[string][]byte, modes map[string]int64) {
	var (
		tr  = tar.NewReader(reader)
		hdr *tar.Header
		buf []byte
		err error
	)

	content, modes = make(map[string][]byte), make(map[string]int64)
	for {
		if hdr, err = tr.Next(); errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("tar.Next() вернул ошибку: %v", err)
		}
		if buf, err = io.ReadAll(tr); err != nil {
			t.Fatalf("tar.ReadAll() вернул ошибку: %v", err)
		}
		content[hdr.Name] = buf
		modes[hdr.Name] = hdr.Mode
	}

	return
}

func TestImplementationConstructor(t *testing.T) {
	var impl = Constructor()

	if impl == nil {
		t.Fatal("Constructor() вернул nil")
	}
	if impl.Res == nil {
		t.Fatal("Constructor() не создал карту ресурсов")
	}
	if len(impl.Group()) != 0 {
		t.Fatal("новый объект не должен содержать групп ресурсов")
	}
}

func TestImplementationAddAndResourceData(t *testing.T) {
	var (
		impl    = Constructor()
		content = []byte{0x01, 0x02, 0x03, 0x00}
		res     = testResource(content)
	)

	if err := impl.Add("assets", "bin/data.bin", res); err != nil {
		t.Fatalf("Add() вернул ошибку: %v", err)
	}
	got := impl.ResourceData("assets", "bin/data.bin")
	if got == nil {
		t.Fatal("ResourceData() вернул nil после Add()")
	}
	if got == &res {
		t.Fatal("ResourceData() вернул исходный объект ресурса, ожидалась копия")
	}
	if got.Size != res.Size || !got.Time.Equal(res.Time) || got.ContentType != res.ContentType {
		t.Fatalf("ResourceData() вернул описание %+v, ожидалось %+v", got, res)
	}
	if !bytes.Equal(got.Content, content) {
		t.Fatalf("ResourceData().Content = %v, ожидалось %v", got.Content, content)
	}
	res.Content[0] = 0xff
	if stored := impl.ResourceData("assets", "bin/data.bin"); bytes.Equal(stored.Content, res.Content) {
		t.Fatal("Add() не скопировал контент ресурса")
	}
	got.Content[1] = 0xff
	if stored := impl.ResourceData("assets", "bin/data.bin"); stored.Content[1] == 0xff {
		t.Fatal("ResourceData() вернул контент без копирования")
	}
	if groups := impl.Group(); len(groups) != 1 || groups[0] != "assets" {
		t.Fatalf("Group() = %v, ожидалась группа assets", groups)
	}
	if names := impl.ResourceByGroup("assets"); len(names) != 1 || names[0] != "bin/data.bin" {
		t.Fatalf("ResourceByGroup() = %v, ожидался ресурс bin/data.bin", names)
	}
	if impl.ResourceData("unknown", "bin/data.bin") != nil {
		t.Fatal("ResourceData() для неизвестной группы вернул не nil")
	}
	if impl.ResourceData("assets", "unknown") != nil {
		t.Fatal("ResourceData() для неизвестного ресурса вернул не nil")
	}
}

func TestImplementationAddErrors(t *testing.T) {
	var (
		impl = Constructor()
		res  = testResource([]byte("valid"))
	)

	if err := impl.Add("assets", "file.txt", res); err != nil {
		t.Fatalf("Add() вернул ошибку: %v", err)
	}
	if err := impl.Add("assets", "file.txt", res); err == nil {
		t.Fatal("Add() при дублировании ресурса вернул nil ошибку")
	}
	//goland:noinspection GoMaybeNil
	if stored := impl.ResourceData("assets", "file.txt"); !bytes.Equal(stored.Content, res.Content) {
		t.Fatal("дублирующий Add() изменил существующий ресурс")
	}
	brokenNewGroup := testResource([]byte("123"))
	brokenNewGroup.Size = 10
	if err := impl.Add("broken", "bad-size", brokenNewGroup); err == nil {
		t.Fatal("Add() с неверным размером вернул nil ошибку")
	}
	if got := impl.ResourceData("broken", "bad-size"); got != nil {
		t.Fatalf("Add() с неверным размером оставил ресурс: %+v", got)
	}
	if names := impl.ResourceByGroup("broken"); len(names) != 0 {
		t.Fatalf("Add() с неверным размером создал группу с ресурсами: %v", names)
	}
	broken := testResource([]byte("123"))
	broken.Size = 10
	if err := impl.Add("assets", "bad-size", broken); err == nil {
		t.Fatal("Add() с неверным размером в существующей группе вернул nil ошибку")
	}
	if got := impl.ResourceData("assets", "bad-size"); got != nil {
		t.Fatalf("Add() с неверным размером в существующей группе оставил ресурс: %+v", got)
	}
}

func TestImplementationNilResourceData(t *testing.T) {
	var impl = Constructor()

	impl.Res["nil"] = map[string]*Resource{"item": nil}
	if got := impl.ResourceData("nil", "item"); got != nil {
		t.Fatalf("ResourceData() для nil ресурса вернул %+v", got)
	}
	reader, err := impl.ResourceByGroupTarReader("nil")
	if err != nil {
		t.Fatalf("ResourceByGroupTarReader() для nil ресурса вернул ошибку: %v", err)
	}
	content, _ := readTarFiles(t, reader)
	if len(content) != 0 {
		t.Fatalf("ResourceByGroupTarReader() для nil ресурса вернул файлы: %v", content)
	}
}

func TestImplementationResourceByGroupTarReader(t *testing.T) {
	var (
		impl     = Constructor()
		text     = []byte("hello")
		bin      = []byte{0x00, 0xff, 0x10}
		empty    []byte
		emptyRes = testResource(empty)
	)

	emptyRes.ContentType = "text/plain"
	for _, item := range []struct {
		name string
		res  Resource
	}{
		{name: "a.txt", res: testResource(text)},
		{name: "dir/b.bin", res: testResource(bin)},
		{name: "empty.txt", res: emptyRes},
	} {
		if err := impl.Add("assets", item.name, item.res); err != nil {
			t.Fatalf("Add(%q) вернул ошибку: %v", item.name, err)
		}
	}
	reader, err := impl.ResourceByGroupTarReader("assets")
	if err != nil {
		t.Fatalf("ResourceByGroupTarReader() вернул ошибку: %v", err)
	}
	if reader == nil {
		t.Fatal("ResourceByGroupTarReader() вернул nil reader")
	}
	content, modes := readTarFiles(t, reader)
	if len(content) != 3 {
		t.Fatalf("tar контейнер содержит %d файла, ожидалось 3: %v", len(content), content)
	}
	if !bytes.Equal(content["a.txt"], text) || !bytes.Equal(content["dir/b.bin"], bin) || !bytes.Equal(content["empty.txt"], empty) {
		t.Fatalf("tar контейнер содержит неверный контент: %v", content)
	}
	for name, mode := range modes {
		if mode != 0600 {
			t.Fatalf("mode файла %q = %o, ожидался 0600", name, mode)
		}
	}
	reader, err = impl.ResourceByGroupTarReader("unknown")
	if err != nil {
		t.Fatalf("ResourceByGroupTarReader() для неизвестной группы вернул ошибку: %v", err)
	}
	content, _ = readTarFiles(t, reader)
	if len(content) != 0 {
		t.Fatalf("tar контейнер для неизвестной группы содержит файлы: %v", content)
	}
}

func TestImplementationResourceByGroupTarReaderErrors(t *testing.T) {
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
			var impl = Constructor()
			impl.Res["assets"] = map[string]*Resource{
				"broken": {
					Size:    test.size,
					Time:    time.Now(),
					Content: test.content,
				},
			}
			reader, err := impl.ResourceByGroupTarReader("assets")
			if err == nil {
				t.Fatal("ResourceByGroupTarReader() вернул nil ошибку")
			}
			if reader != nil {
				t.Fatal("ResourceByGroupTarReader() при ошибке вернул не nil reader")
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("ResourceByGroupTarReader() вернул ошибку %q, ожидалось %q", err.Error(), test.wantErr)
			}
		})
	}
}

func TestImplementationConcurrentAccess(t *testing.T) {
	var (
		err   error
		n, m  int
		impl  = Constructor()
		wg    sync.WaitGroup
		errCh = make(chan error, 128)
	)

	for n = range 8 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for m = range 20 {
				var (
					name    = fmt.Sprintf("dir/%d/%d.txt", n, m)
					content = []byte(name)
					err     error
				)
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
		t.Fatalf("concurrent access вернул ошибку: %v", err)
	}
}
