package resource

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"sort"
)

// Constructor Конструктор сущности пакета.
func Constructor() *Implementation {
	var r6e = &Implementation{Res: make(map[string]map[string]*Resource)}
	return r6e
}

// Add Добавление ресурса в группу ресурсов.
func (r6e *Implementation) Add(group string, name string, resource Resource) (err error) {
	const (
		errGroupExists  = "ресурс %q в группе ресурсов %q уже существует"
		errResourceSize = "размер контента в описании не соответствует фактическому размеру контента"
	)
	var (
		items       map[string]*Resource
		groupExists bool
	)

	r6e.ResLock.Lock()
	defer r6e.ResLock.Unlock()
	if items, groupExists = r6e.Res[group]; groupExists {
		if _, ok := items[name]; ok {
			err = fmt.Errorf(errGroupExists, name, group)
			return
		}
	}
	if uint64(len(resource.Content)) != resource.Size {
		err = errors.New(errResourceSize)
		return
	}
	if !groupExists {
		items = make(map[string]*Resource)
		r6e.Res[group] = items
	}
	items[name] = resourceCopy(&resource)

	return
}

// Group Список групп ресурсов.
func (r6e *Implementation) Group() (ret []string) {
	var group string

	r6e.ResLock.RLock()
	defer r6e.ResLock.RUnlock()
	ret = make([]string, 0, len(r6e.Res))
	for group = range r6e.Res {
		ret = append(ret, group)
	}

	return
}

// ResourceByGroup Список ресурсов в указанной группе.
func (r6e *Implementation) ResourceByGroup(group string) (ret []string) {
	var (
		rn string
		ok bool
	)

	r6e.ResLock.RLock()
	defer r6e.ResLock.RUnlock()
	if _, ok = r6e.Res[group]; !ok {
		return
	}
	ret = make([]string, 0, len(r6e.Res[group]))
	for rn = range r6e.Res[group] {
		ret = append(ret, rn)
	}

	return
}

// ResourceData Получение ресурса по имени группы и ресурсу.
// Если ресурса нет, возвращается nil.
func (r6e *Implementation) ResourceData(group string, resource string) (ret *Resource) {
	var (
		item *Resource
		ok   bool
	)

	r6e.ResLock.RLock()
	defer r6e.ResLock.RUnlock()
	if _, ok = r6e.Res[group]; !ok {
		return
	}
	if item, ok = r6e.Res[group][resource]; !ok || item == nil {
		return
	}
	ret = resourceCopy(item)

	return
}

// ResourceByGroupTarReader В памяти создаётся tar контейнер со всеми ресурсами группы и
// возвращается *bytes.Reader к tar контейнеру.
func (r6e *Implementation) ResourceByGroupTarReader(group string) (ret *bytes.Reader, err error) {
	const (
		errHeaderWrite = "запись заголовка файла %q прервана ошибкой: %s"
		errBodyWrite   = "запись тела файла %q прервана ошибкой: %s"
		errCreateClose = "создание tar контейнера, в момент закрытия, прервано ошибкой: %s"
	)
	type tarSource struct {
		name string
		data *Resource
	}
	var (
		buf    *bytes.Buffer
		th     *tar.Header
		tw     *tar.Writer
		source []tarSource
		items  map[string]*Resource
		ok     bool
		src    tarSource
	)

	r6e.ResLock.RLock()
	items, ok = r6e.Res[group]
	source = make([]tarSource, 0, len(items))
	if ok {
		for name, data := range items {
			if data == nil {
				continue
			}
			source = append(source, tarSource{name: name, data: resourceCopy(data)})
		}
	}
	r6e.ResLock.RUnlock()
	sort.Slice(source, func(i, j int) bool { return source[i].name < source[j].name })
	buf = new(bytes.Buffer)
	//goland:noinspection GoResourceLeak
	tw = tar.NewWriter(buf)
	for _, src = range source {
		// Заголовок файла в tar контейнере.
		th = &tar.Header{
			Name:       src.name,
			Mode:       0600,
			Size:       int64(src.data.Size),
			ModTime:    src.data.Time,
			AccessTime: src.data.Time,
			ChangeTime: src.data.Time,
		}
		// Запись заголовка файла в tar контейнер.
		if err = tw.WriteHeader(th); err != nil {
			err = fmt.Errorf(errHeaderWrite, src.name, err)
			return
		}
		// Запись тела файла в tar контейнер.
		if _, err = tw.Write(src.data.Content); err != nil {
			err = fmt.Errorf(errBodyWrite, src.name, err)
			return
		}
	}
	if err = tw.Close(); err != nil {
		err = fmt.Errorf(errCreateClose, err)
		return
	}
	ret = bytes.NewReader(buf.Bytes())

	return
}

// resourceCopy Создание копии описания ресурса.
func resourceCopy(src *Resource) (ret *Resource) {
	ret = &Resource{
		Size:        src.Size,
		Time:        src.Time,
		ContentType: src.ContentType,
		Content:     append([]byte(nil), src.Content...),
	}

	return
}
