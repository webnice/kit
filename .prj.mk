## Project tooling makefile

## Сценарий по умолчанию - отображение доступных команд
default: help

## Переопределение названия проекта
define APP
wn
endef
export APP

## Загрузка зависимостей проекта
define PROJECT_DEPENDENCES
  @go clean
  @go get -d ...
endef

## Сборка проекта
define PROJECT_BUILD
	@mkdir -p ${DIR}/pkg 2>/dev/null; true
	@cd "${DIR}/src/$(APP)"; go build \
    -o ${BIN01} \
    -gcflags=-N -gcflags=-l \
    -ldflags "${LDFLAGS}" \
    -pkgdir ${DIR}/pkg \
		-trimpath \
    .
endef
export PROJECT_BUILD

## Запуск приложения в режиме разработки/отладки
define PROJECT_RUN_DEVELOPMENT
	@make clear
	@${BIN01} test test\
	; echo "------ exit=$$?"
endef
export PROJECT_RUN_DEVELOPMENT

## Запуск приложения в продакшн режиме
define PROJECT_RUN_PRODUCTION
	@${BIN01} daemon
endef
export PROJECT_RUN_PRODUCTION

## Тестовая команда, которую можно запустить обычным вызовом: make vvv
cmdtest:
	@echo "cmdtest: run 'bin/$(APP) -d version'"
	@${BIN01} -d version
.PHONY: cmdtest

## Помощь по дополнительным командам
define PROJECT_HELP
  @#    "    команда              - Описание команды"
	@echo "    cmdtest              - Тестовая команда описанная в подключаемом файле .prj.mk"
endef
export PROJECT_HELP
