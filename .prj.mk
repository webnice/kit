## Project tooling makefile

## Сценарий по умолчанию - отображение доступных команд
default: help

## Переопределение названия проекта
define APP
wn
endef
export APP

## Основной исполняемый файл проекта.
BIN01 := $(DIR)/bin/$(APP)
## Файл с идентификатором процесса запущенного приложения.
PIDF01 := $(DIR)/run/$(APP).pid
PIDN01 = $(shell if [ -f "$(PIDF01)" ]; then cat "$(PIDF01)"; fi)

## Загрузка зависимостей проекта, продакшн режим.
define PROJECT_DEPENDENCIES
  @#go get -d ...
endef

## Загрузка зависимостей проекта, режим разработки.
define PROJECT_DEPENDENCIES_DEVELOPMENT
	@#go get -d ...
endef

## Создание дополнительных директорий проекта, вызывается в make dep
define PROJECT_FOLDERS
	@mkdir -p ${DIR}/bin; true
	@mkdir -p ${DIR}/pkg; true
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
	@echo "    build                - Компиляция приложения."
	@echo "    dev                  - Запуск приложения в режиме разработки."
	@echo "    run                  - Запуск приложения в продакшн режиме."
	@echo "    kill                 - Принудительное завершение процесса приложения."
	@echo "    version              - Вывод на экран версии приложения."
	@echo "    cmdtest              - Тестовая команда описанная в подключаемом файле .prj.mk"
endef
export PROJECT_HELP

## Встраивание статических ресурсов в приложение.
embed:
.PHONY: embed

## Компиляция проекта.
build:
	@mkdir -p "$(DIR)/bin" "$(DIR)/pkg"; true
	$(call PROJECT_BUILD)
.PHONY: build

## Запуск приложения в режиме разработки.
dev: clear
	$(call PROJECT_RUN_DEVELOPMENT)
.PHONY: dev

## Запуск приложения в продакшн режиме.
run:
	$(call PROJECT_RUN_PRODUCTION)
.PHONY: run

## Принудительное завершение процесса запущенного приложения.
kill:
	@if [ ! "$(PIDN01)x" = "x" ]; then \
		sudo kill -KILL "$(PIDN01)" 2>/dev/null; \
		if [ $$? -ne 0 ]; then \
			echo "Нет такого процесса: $(PIDN01)"; \
		else \
			rm "$(PIDF01)" 2>/dev/null; true; \
		fi; \
	else \
		echo "Файл '$(PIDF01)' отсутствует или пуст."; \
	fi
.PHONY: kill

## Вывод на экран версии приложения.
version: v
.PHONY: version
v:
	@"$(BIN01)" version 2>/dev/null; true
.PHONY: v

## Сборка RPM пакета.
rpm:
.PHONY: rpm

## Создание директорий миграций.
migration-mkdir:
.PHONY: migration-mkdir

## Команды инструментов миграций.
migration-commands:
.PHONY: migration-commands
