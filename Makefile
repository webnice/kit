## Project tooling makefile.
## Version: 12.05.2022.

## Вычисление текущей директории проекта.
export DIR   := $(strip $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))))
## Определение директории размещения бинарных файлов.
export GOBIN := $(DIR)/bin
## Подключение переменных окружения проекта.
ifneq ("$(wildcard $(DIR)/.env.mk)","")
include $(DIR)/.env.mk
endif
## Подключение makefile части содержащей специфичные для проекта команды.
ifneq ("$(wildcard $(DIR)/.prj.mk)","")
include $(DIR)/.prj.mk
endif

APP                 ?= $(shell basename $(DIR))
GOPATH              := $(GOPATH)
DATE                := $(shell date -u +%Y%m%d.%H%M%S.%Z)
LDFLAGS              = -X main.build=$(DATE) $(PROJECT_LDFLAGS:'')
GOGENERATE           = $(shell if [ -f .gogenerate ]; then cat .gogenerate; fi)
TESTPACKETS          = $(shell if [ -f .testpackages ]; then cat .testpackages; fi)
BENCHPACKETS         = $(shell if [ -f .benchpackages ]; then cat .benchpackages; fi)
LOCALPACKAGES        = $(shell if [ -f .localpackages ]; then cat .localpackages; fi)

BIN01               := $(DIR)/bin/$(APP)
BINVERSION          := $(shell ${BIN01} version 2>/dev/null; true)
VERN01              := $(shell echo "$(BINVERSION)" | awk -F '-' '{ print $$1 }' )
VERB01              := $(shell echo "$(BINVERSION)" | awk -F 'build.' '{ print $$2 }' )
PIDF01              := $(DIR)/run/$(APP).pid
PIDN01               = $(shell if [ -f $(PIDF01) ]; then cat $(PIDF01); fi)

PROJECT_CGO_ENABLED ?= $(PROJECT_CGO_ENABLED:0)

## Сценарий по умолчанию - отображение доступных команд.
default: help

## Загрузка зависимостей.
dep-init:
	@rm -rf ${DIR}/vendor 2>/dev/null; true
	$(call PROJECT_FOLDERS)
.PHONY: dep-init
dep: dep-init
	@for item in $(LOCALPACKAGES); do PKGNAME=`echo $${item} | awk -F'=' '{print $$1}'`; REPLACE=`echo $${item} | awk -F'=' '{print $$2}'`; \
		go mod edit -dropreplace $${PKGNAME}; \
	done
	@go clean -cache -modcache
	@go get -u -v ./...
	@go mod download
	@go mod tidy
	@go mod vendor
	$(call PROJECT_DEPENDENCES)
.PHONY: dep
dep-dev: dep-init
	@for item in $(LOCALPACKAGES); do PKGNAME=`echo $${item} | awk -F'=' '{print $$1}'`; REPLACE=`echo $${item} | awk -F'=' '{print $$2}'`; \
		go mod edit -replace $${PKGNAME}=$${REPLACE}; \
	done
	@go clean -cache -modcache
	@go get -u -v ./...
	@go mod download
	@go mod tidy
	$(call PROJECT_DEPENDENCES_DEVELOPMENT)
.PHONY: dep-dev

## Кодогенерация (run only during development).
## All generating files are included in a .gogenerate file.
gen: dep-init
	@for PKGNAME in $(GOGENERATE); do go generate $${PKGNAME}; done
.PHONY: gen

## Project building for environment architecture.
build:
	$(call PROJECT_BUILD)
.PHONY: build

## Run application in development mode.
dev: clear
	@$(call PROJECT_RUN_DEVELOPMENT)
.PHONY: dev

## Run application in production mode.
run:
	$(call PROJECT_RUN_PRODUCTION)
.PHONY: run

## Kill process and remove pid file.
kill:
	@if [ ! "$(PIDN01)x" == "x" ]; then \
		sudo kill -KILL "$(PIDN01)" 2>/dev/null; \
		if [ $$? -ne 0 ]; then echo "No such process ID: $(PIDN01)"; \
		else rm "$(PIDF01)" 2>/dev/null; true; fi; \
	fi
.PHONY: kill

## Getting application version.
version: v
v:
	@${BIN01} version 2>/dev/null
.PHONY: v
.PHONY: version

## RPM build openSUSE linux version.
RPMBUILD_OS ?= $(RPMBUILD_OS:leap)
RPMBUILD_OS ?= $(RPMBUILD_OS:tumbleweed)

## Migration tools for all databases.
## Please see files .env and .env_example, for setup access to databases.
####################################
COMMANDS  = up create down status redo version
MTARGETS := $(shell \
for cmd in $(COMMANDS); do \
	for drv in $(MIGRATIONS); do \
		echo "m-$${drv}-$${cmd}"; \
	done; \
done)
## Migration tools create directory.
migration-mkdir:
	@for dir in $$(echo $(MIGRATIONS)); do \
		mkdir -p "$(DIR)/migrations/$${dir}"; true; \
	done
.PHONY: migration-mkdir
## Migration tools gets data from env.
MIGRATION_DIR  = ${$(shell echo $(shell echo "${@}" | sed -e 's/^m-\(.*\)-\(.*\)$$/\1/') | awk '{print "GOOSE_DIR_"toupper($$0)}')}
MIGRATION_DRV  = ${$(shell echo $(shell echo "${@}" | sed -e 's/^m-\(.*\)-\(.*\)$$/\1/') | awk '{print "GOOSE_DRV_"toupper($$0)}')}
MIGRATION_DSN  = ${$(shell echo $(shell echo "${@}" | sed -e 's/^m-\(.*\)-\(.*\)$$/\1/') | awk '{print "GOOSE_DSN_"toupper($$0)}')}
MIGRATION_CMD  = $(shell echo $(shell echo "${@}" | sed -e 's/^m-\(.*\)-\(.*\)$$/\2/'))
MIGRATION_TMP := $(shell mktemp)
## Migration tools targets.
migration-commands: $(MTARGETS)
$(MTARGETS): migration-mkdir
	@if [ "$(MIGRATION_CMD)" == "create" ]; then\
		read -p "Введите название миграции: " MGRNAME && \
		if [ "$${MGRNAME}" == "" ]; then MGRNAME="new"; fi && \
		echo "$${MGRNAME}" > "$(MIGRATION_TMP)"; \
	fi
	@if ([ ! "`cat $(MIGRATION_TMP)`" = "" ]) && ([ "$(MIGRATION_CMD)" == "create" ]); then\
		GOOSE_DIR="$(MIGRATION_DIR)" GOOSE_DRV="$(MIGRATION_DRV)" GOOSE_DSN="$(MIGRATION_DSN)" gsmigrate $(MIGRATION_CMD) "`cat $(MIGRATION_TMP)`"; \
	else \
		GOOSE_DIR="$(MIGRATION_DIR)" GOOSE_DRV="$(MIGRATION_DRV)" GOOSE_DSN="$(MIGRATION_DSN)" gsmigrate $(MIGRATION_CMD); \
	fi
	@if [ -f "$(MIGRATION_TMP)" ]; then rm "$(MIGRATION_TMP)"; fi
.PHONY: migration-commands $(MTARGETS)
####################################

## Testing one or multiple packages as well as applications with reporting on the percentage of test coverage.
## All testing files are included in a .testpackages file.
# test:
# 	@echo "mode: set" > $(DIR)/log/coverage.log
# 	@for PACKET in $(TESTPACKETS); do \
# 		touch coverage-tmp.log; \
# 		GOPATH=${GOPATH} go test -v -covermode=count -coverprofile=$(DIR)/log/coverage-tmp.log $$PACKET; \
# 		if [ "$$?" -ne "0" ]; then exit $$?; fi; \
# 		tail -n +2 $(DIR)/log/coverage-tmp.log | sort -r | awk '{if($$1 != last) {print $$0;last=$$1}}' >> $(DIR)/log/coverage.log; \
# 		rm -f $(DIR)/log/coverage-tmp.log; true; \
# 	done
# .PHONY: test

## Displaying in the browser coverage of tested code, on the html report (run only during development).
# cover: test
# 	@GOPATH=${GOPATH} go tool cover -html=$(DIR)/log/coverage.log
# .PHONY: cover

## Performance testing.
## All testing files are included in a .benchpackages file.
# bench:
# 	@for PACKET in $(BENCHPACKETS); do GOPATH=${GOPATH} go test -race -bench=. -benchmem $$PACKET; done
# .PHONY: bench

## Code quality testing.
# lint:
# 	$(call PROJECT_LINTER)
# .PHONY: lint

## Cleaning console screen.
clear:
	clear
.PHONY: clear

## Clearing project temporary files.
clean:
	@go clean -cache -modcache
	@chown -R `whoami` ${DIR}/pkg/; true
	@chmod -R 0777 ${DIR}/pkg/; true
	@rm -rf ${DIR}/bin/*; true
	@rm -rf ${DIR}/pkg/*; true
	@rm -rf ${DIR}/run/*.pid; true
	@rm -rf ${DIR}/log/*.log; true
	@rm -rf ${DIR}/rpmbuild; true
	@rm -rf ${DIR}/*.log; true
	@export DIR=
.PHONY: clean

## Help for main targets.
help:
	@echo "Usage: make [target]"
	@echo "  target is:"
	@echo "    dep                  - Загрузка и обновление зависимостей проекта."
	@echo "    dep-dev              - Загрузка и обновление зависимостей проекта для среды разработки."
	@#echo "    gen                  - Кодогенерация с использованием go generate."
	@echo "    build                - Компиляция приложения."
	@#echo "    build-i386           - Компиляция приложения для архитектуры i386."
	@echo "    run                  - Запуск приложения в продакшн режиме."
	@echo "    dev                  - Запуск приложения в режиме разработки."
	@echo "    kill                 - Отправка приложению сигнала kill -HUP, используется в случае зависания."
	@echo "    m-[driver]-[command] - Работа с миграциями базы данных."
	@echo "                           Используемые базы данных (driver) описываются в файле .env."
	@echo "                           Доступные драйверы баз данных: mysql clickhouse sqlite3 postgres redshift tidb."
	@echo "                           Доступные команды: up, down, create, status, redo, version."
	@echo "                           Пример команд при включённой базе данных mysql:"
	@echo "                             make m-mysql-up      - Примернение миграций до самой последней версии."
	@echo "                             make m-mysql-down    - Отмена последней миграции."
	@echo "                             make m-mysql-create  - Создание нового файла миграции."
	@echo "                             make m-mysql-status  - Статус всех миграций базы данных."
	@echo "                             make m-mysql-redo    - Отмена и повторное применение последней миграции."
	@echo "                             make m-mysql-version - Отображение версии базы данных (применённой миграции)."
	@echo "                           Подробная информаци по командам доступна в документации утилиты gsmigrate."
	@echo "    version              - Вывод на экран версии приложения."
	@#echo "    rpm                  - Создание RPM пакета."
	@#echo "    rpm-i386             - Создание RPM пакета для архитектуры i386."
	@#echo "    bench                - Запуск тестов производительности проекта."
	@#echo "    test                 - Запуск тестов проекта."
	@#echo "    cover                - Запуск тестов проекта с отображением процента покрытия кода тестами."
	@#echo "    lint                 - Запуск проверки кода с помощью gometalinter."
	@echo "    clean                - Очистка директории проекта от временных файлов."
	@$(call PROJECT_HELP)
.PHONY: help
