package resource

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
)

// Constructor Конструктор сущности пакета.
func Constructor() *Implementation {
	var r6e = &Implementation{Res: make(map[string]map[string]*Resource)}
	return r6e
}

// Add Добавление ресурса в группу ресурсов.
func (r6e *Implementation) Add(group string, name string, resource Resource) (err error) {
	var (
		ok bool
		n  int
	)

	r6e.ResLock.Lock()
	defer r6e.ResLock.Unlock()
	if _, ok = r6e.Res[group]; !ok {
		r6e.Res[group] = make(map[string]*Resource)
	}
	if _, ok = r6e.Res[group][name]; ok {
		err = fmt.Errorf("ресурс %q в группе ресурсов %q уже существует", name, group)
		return
	}
	r6e.Res[group][name] = &Resource{
		Size:        resource.Size,
		Time:        resource.Time,
		ContentType: resource.ContentType,
		Content:     make([]byte, len(resource.Content)),
	}
	n = copy(r6e.Res[group][name].Content, resource.Content)
	if uint64(n) != r6e.Res[group][name].Size {
		err = errors.New("размер контента в описании не соответствует фактическому размеру контента")
		return
	}

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
	var ok bool

	r6e.ResLock.RLock()
	defer r6e.ResLock.RUnlock()
	if _, ok = r6e.Res[group]; !ok {
		return
	}
	if _, ok = r6e.Res[group][resource]; !ok {
		return
	}
	ret = &Resource{
		Size:        r6e.Res[group][resource].Size,
		Time:        r6e.Res[group][resource].Time,
		ContentType: r6e.Res[group][resource].ContentType,
		Content:     make([]byte, len(r6e.Res[group][resource].Content)),
	}
	copy(ret.Content, r6e.Res[group][resource].Content)

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
	var (
		buf  *bytes.Buffer
		name string
		th   *tar.Header
		tw   *tar.Writer
		data *Resource
	)

	buf = new(bytes.Buffer)
	tw = tar.NewWriter(buf)
	for _, name = range r6e.ResourceByGroup(group) {
		// Получение даты и времени модификации самого свежего файла.
		data = r6e.ResourceData(group, name)
		// Заголовок файла в tar контейнере.
		th = &tar.Header{
			Name:       name,
			Mode:       0600,
			Size:       int64(data.Size),
			ModTime:    data.Time,
			AccessTime: data.Time,
			ChangeTime: data.Time,
		}
		// Запись заголовка файла в tar контейнер.
		if err = tw.WriteHeader(th); err != nil {
			err = fmt.Errorf(errHeaderWrite, name, err)
			return
		}
		// Запись тела файла в tar контейнер.
		if _, err = tw.Write(data.Content); err != nil {
			err = fmt.Errorf(errBodyWrite, name, err)
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
