package server

import (
	"fmt"
	"reflect"
	runtimeDebug "runtime/debug"

	kitModuleUuid "github.com/webnice/kit/v4/module/uuid"
	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

// Поиск конфигураций серверов по типу и UUID, если список IDs пустой, возврат всех серверов указанного типа.
func (sio *impl[T]) serverGetByTypeById(t interface{}, IDs ...string) (ret []*server[master], err error) {
	var (
		ok   bool
		n, s int
		uus  []kitModuleUuid.UUID
		srv  *server[master]
		kts  kitTypesServer.Type
	)

	// Выбор типа.
	switch t.(type) {
	case *master:
		return
	case *kitTypesServer.Web:
		kts = kitTypesServer.TWeb
	case *kitTypesServer.Grpc:
		kts = kitTypesServer.TGrpc
	case *kitTypesServer.Tcp:
		kts = kitTypesServer.TTcp
	default:
		err = sio.Errors().TypeNotImplemented(reflect.TypeOf(t).String())
		return
	}
	// Проверка существования sWeb конфигураций.
	sio.serverLock.RLock()
	_, ok = sio.server[kts]
	sio.serverLock.RUnlock()
	if !ok {
		return
	}
	// Обработка запроса всех конфигураций серверов.
	if len(IDs) == 0 {
		sio.serverLock.RLock()
		ret = make([]*server[master], 0, len(sio.server[kts]))
		for s = range sio.server[kts] {
			server := (interface{}(*sio.server[kts][s])).(server[master])
			ret = append(ret, &server)
		}
		sio.serverLock.RUnlock()
		return
	}
	// Парсинг UUID идентификаторов.
	if uus, err = sio.parseIDsToObject(IDs...); err != nil {
		return
	}
	// Выборка конфигураций серверов по UUID.
	sio.serverLock.RLock()
	defer sio.serverLock.RUnlock()
	// Выделение памяти для результата.
	ret = make([]*server[master], 0, len(IDs))
	for n = range uus {
		srv = nil
		for s = range sio.server[kts] {
			if sio.server[kts][s].s.ID.Equal(uus[n]) {
				server := (interface{}(*sio.server[kts][s])).(server[master])
				srv = &server
				break
			}
		}
		if srv == nil {
			err = sio.Errors().ServerWithUuidNotFound(IDs[n])
			return
		}
		ret = append(ret, srv)
	}

	return
}

// Конвертирование UUID из среза строк в срез объектов.
func (sio *impl[T]) parseIDsToObject(IDs ...string) (ret []kitModuleUuid.UUID, err error) {
	var (
		uuo kitModuleUuid.UUID
		n   int
	)

	// Парсинг UUID идентификаторов.
	ret = make([]kitModuleUuid.UUID, 0, len(IDs))
	for n = range IDs {
		if uuo = kitModuleUuid.Get().FromString(IDs[n]); uuo.Equal(kitModuleUuid.NULL) {
			err = sio.Errors().UuidError(IDs[n])
			return
		}
		ret = append(ret, uuo)
	}

	return
}

// Запуск потока, ожидание ошибки из канала, возвращение ошибки.
// Наличие ошибки означает что запущенная горутина остановилась или не была запущена.
// Ошибка равная nil сигнализирует об успешном запуске потока.
func goWaitError(f func(chan<- error)) (err error) {
	var (
		onStart chan error
	)

	defer func() {
		if e := recover(); e != nil {
			switch et := e.(type) {
			case error:
				err = et
			default:
				err = fmt.Errorf("%v", e)
			}
			err = fmt.Errorf("%s\n%s", err, string(runtimeDebug.Stack()))
		}
	}()
	onStart = make(chan error)
	go f(onStart)
	err = safeWaitError(onStart)

	return
}

// Ожидание ошибки из канала, закрытие канала получения ошибки.
func safeWaitError(ch chan error) (err error) {
	defer func() { _ = recover() }()
	err = <-ch
	close(ch)

	return
}
