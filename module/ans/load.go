package ans

import (
	"net/http"

	"github.com/webnice/dic"
	kitModuleRqVar "github.com/webnice/kit/v4/module/rqvar"
)

// RqVar Интерфейс модуля github.com/webnice/kit/module/rqvar. Поиск и загрузка данных из:
// + контекста HTTP запроса;
// + заголовков HTTP запроса;
// + "печенек" HTTP запроса;
// + параметров HTTP запроса, а именно в URL Request;
// + путей URN роутинга HTTP запроса;
// + вызов функции структуры загружаемых данных;
func (ans *impl) RqVar() kitModuleRqVar.Interface { return kitModuleRqVar.Get() }

// RqLoad Выполнение загрузки и декодирования данных, в случае возникновения ошибки формируется и отправляется HTTP
// ответ с кодом 400, содержащий возникшую ошибку.
// Данные загружаются из тела запроса в переменную variable с использованием декодирования данных, выбор кодера
// осуществляется на основе заголовка запроса Content-Type, поддерживаются два метода сериализации данных: JSON, XML.
// Ответ с ошибкой сериализуется тем же самым методом сериализации данных, что был в запросе.
func (ans *impl) RqLoad(wr http.ResponseWriter, rq *http.Request, variable any) (err error) {
	if err = ans.RqData(rq, variable); err != nil {
		ans.NewRestError(dic.Status().BadRequest, err).Json(wr)
		return
	}

	return
}

// RqLoadVerify Выполнение загрузки и декодирования данных, в случае возникновения ошибки формируется и
// отправляется HTTP ответ с кодом 400, содержащий возникшую ошибку.
// Загруженные данные проверяются библиотекой github.com/go-playground/validator.
// Данные загружаются из тела запроса в переменную variable с использованием декодирования данных, выбор кодера
// осуществляется на основе заголовка запроса Content-Type, поддерживаются два метода сериализации данных: JSON, XML.
// Ответ с ошибкой сериализуется тем же самым методом сериализации данных, что был в запросе.
func (ans *impl) RqLoadVerify(wr http.ResponseWriter, rq *http.Request, variable any) (err error) {
	var reo RestErrorInterface

	if err = ans.RqLoad(wr, rq, variable); err != nil {
		return
	}
	if reo = ans.Verify(variable); reo == nil {
		return
	}
	err = reo.AsError()
	reo.Json(wr)

	return
}
