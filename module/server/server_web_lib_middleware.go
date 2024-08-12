package server

const (
	// Название контекста для объекта контроля ВЕБ сервера.
	contextNameMiddlewareWebServerControl = "middleware-web-server-control"
)

var (
	// Ключ контекста для объекта контроля ВЕБ сервера.
	contextKeyMiddlewareWebServerControl = &contextKey{contextNameMiddlewareWebServerControl}
)
