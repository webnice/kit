package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/webnice/debug"
	"github.com/webnice/dic"
	kitTypesServer "github.com/webnice/kit/v4/types/server"
	"github.com/webnice/web/v4"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// IsStarted Состояние выполнения сервера.
func (srv *server[T]) IsStarted() bool { return srv.isStarted }

// Процесс сервера, данный процесс обслуживает один сокет/порт/конфигурацию сервера и выполняется как горутина.
func (srv *server[T]) do(onStart chan<- error) {
	if srv.IsStarted() {
		return
	}
	srv.isStarted, srv.isShutdown = true, false
	defer func() { srv.isStarted, srv.isShutdown = false, false }()
	switch srv.t.(type) {
	case *kitTypesServer.Web:
		srv.doWeb(onStart)
	case *kitTypesServer.Grpc:
		srv.doGrpc(onStart)
	case *kitTypesServer.Tcp:
		srv.doTcp(onStart)
	default:
		onStart <- fmt.Errorf("запуск сервера для типа %s не реализован", reflect.TypeOf(srv.t))
		return
	}
}

// Процесс ВЕБ сервера.
func (srv *server[T]) doWeb(onStart chan<- error) {
	var wsv web.Interface

	srv.p.log().Debug(debug.DumperString(
		srv.t,
		len(srv.r),
		"grpc.ResourceCount()", srv.p.Grpc().ResourceCount(),
		"web.ResourceCount()", srv.p.Web().ResourceCount(),
		"tcp.ResourceCount()", srv.p.Tcp().ResourceCount(),
	))

	mux := chi.NewRouter()
	middleware.DefaultLogger = middleware.RequestLogger(
		&middleware.DefaultLogFormatter{Logger: log.New(os.Stdout, "rest: ", 0), NoColor: false},
	)
	mux.Use(middleware.Logger)
	mux.Get("/", func(wr http.ResponseWriter, rq *http.Request) {
		wr.Header().Set(dic.Header().ContentType.String(), dic.Mime().TextPlain.String())
		_, _ = wr.Write([]byte("ok"))
	})

	wsv = web.New().
		Handler(mux)

	srv.p.log().Debug(debug.DumperString(srv.s.Web.Server))

	if srv.l, srv.err = wsv.
		NewListener(&srv.s.Web.Server); srv.err != nil {

		srv.p.log().Error(srv.err)

		onStart <- srv.err
		return
	}
	if srv.err = wsv.
		Serve(srv.l).
		Error(); srv.err != nil {

		srv.p.log().Error(srv.err)

		onStart <- srv.err
		return
	}
	// Отпускаем родительский процесс, ожидавший всё это время запуск сервера.
	onStart <- nil
	// Ожидание завершения работы сервера.
	wsv.Wait()
	// Ошибка обрабатывается только в ситуации когда сервер завершился самостоятельно без каких-либо причин на это.
	if err := wsv.Error(); !srv.isShutdown && err != nil {
		srv.err = err
		if srv.p.debug {
			srv.p.log().Errorf("сервер остановился, ошибка: %s", err)
		}
	}
}

// Процесс GRPC сервера.
func (srv *server[T]) doGrpc(onStart chan<- error) {

	onStart <- nil
	select {}

}

// Процесс TCP/IP сервера.
func (srv *server[T]) doTcp(onStart chan<- error) {

	onStart <- nil
	select {}

}

// Остановка сервера путём закрытия слушателя соединения через стандартный метод Close.
func (srv *server[T]) doStop() (err error) {
	if !srv.isStarted {
		return
	}
	srv.isShutdown = srv.isStarted
	err = srv.l.Close()

	return
}
