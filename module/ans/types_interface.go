package ans

import (
	"bytes"
	"net/http"

	"github.com/webnice/dic"
	kitModuleRqVar "github.com/webnice/kit/v4/module/rqvar"
)

// Interface Интерфейс пакета.
type Interface interface {
	// Gist Интерфейс служебных методов.
	Gist() Essence

	// NewRestError Создание объекта интерфейса, стандартного REST ответа с ошибкой для кодов ошибок 4xx и 5xx.
	NewRestError(status dic.IStatus, err error) RestErrorInterface

	// РАБОТА С ДАННЫМИ ЗАПРОСА

	// RqVar Интерфейс модуля github.com/webnice/kit/module/rqvar. Поиск и загрузка данных из:
	// + контекста HTTP запроса;
	// + заголовков HTTP запроса;
	// + "печенек" HTTP запроса;
	// + параметров HTTP запроса, а именно в URL Request;
	// + путей URN роутинга HTTP запроса;
	// + вызов функции структуры загружаемых данных;
	RqVar() kitModuleRqVar.Interface

	// RqBytes Загрузка тела HTTP запроса в виде среза байт и возвращение объекта *bytes.Buffer.
	RqBytes(rq *http.Request) (ret *bytes.Buffer, err error)

	// RqData Выполнение загрузки данных из тела запроса в переменную variable с использованием декодирования
	// данных, выбор кодера осуществляется на основе заголовка запроса Content-Type.
	// Поддерживаются два метода сериализации данных: JSON, XML.
	RqData(rq *http.Request, variable any) (err error)

	// RqLoad Выполнение загрузки и декодирования данных, в случае возникновения ошибки формируется и отправляется HTTP
	// ответ с кодом 400, содержащий возникшую ошибку.
	// Данные загружаются из тела запроса в переменную variable с использованием декодирования данных, выбор кодера
	// осуществляется на основе заголовка запроса Content-Type, поддерживаются два метода сериализации данных: JSON, XML.
	// Ответ с ошибкой сериализуется тем же самым методом сериализации данных, что был в запросе.
	RqLoad(wr http.ResponseWriter, rq *http.Request, variable any) (err error)

	// RqLoadVerify Выполнение загрузки и декодирования данных, в случае возникновения ошибки формируется и
	// отправляется HTTP ответ с кодом 400, содержащий возникшую ошибку.
	// Загруженные данные проверяются библиотекой github.com/go-playground/validator.
	// Данные загружаются из тела запроса в переменную variable с использованием декодирования данных, выбор кодера
	// осуществляется на основе заголовка запроса Content-Type, поддерживаются два метода сериализации данных: JSON, XML.
	// Ответ с ошибкой сериализуется тем же самым методом сериализации данных, что был в запросе.
	RqLoadVerify(wr http.ResponseWriter, rq *http.Request, variable any) (err error)

	// РАБОТА С ДАННЫМИ ОТВЕТА НА ЗАПРОС

	// NoContent Ответ кодом 204 "No Content" без передачи тела сообщения.
	NoContent(wr http.ResponseWriter) Interface

	// InternalServerError Ответ на запрос с кодом ошибки 500 и структурой описывающей ошибку.
	InternalServerError(wr http.ResponseWriter, err error) Interface

	// ContentType Установка типа контента передаваемых данных.
	ContentType(wr http.ResponseWriter, mime dic.IMime) Interface

	// ResponseBytes Ответ с проверкой передачи данных.
	ResponseBytes(wr http.ResponseWriter, status dic.IStatus, data []byte) Interface

	// Response Ответ с проверкой передачи данных.
	Response(wr http.ResponseWriter, status dic.IStatus, buf *bytes.Buffer) Interface

	// Json Ответ на запрос с сериализацией результата в JSON формат.
	Json(wr http.ResponseWriter, status dic.IStatus, obj any) Interface
}
