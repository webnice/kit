package main

import (
	"github.com/webnice/kit/v4/app"
	"github.com/webnice/kit/v4/module/cfg"
)

var (
	// Номер версии сборки приложения.
	// Для указания номера версии сборки, необходимо использовать аргумент -X, в команде компиляции приложения:
	//   -X main.build="версия"
	// Пример 1:
	//   -X main.build=$(date -u +%Y%m%d.%H%M%S.%Z)
	// Пример 2:
	//   -X main.build="20211001.100001"
	// Если аргумент с номером версии не передан, тогда используется значение по умолчанию.
	// Значение по умолчанию: "dev".
	build = `dev`
)

func main() { cfg.RegistrationMain(app.Get().Main).Gist().Version(version, build).Cfg().Gist().App() }
