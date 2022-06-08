// Package configuration
package configuration

import (
	"fmt"
	"strings"
)

// Create Создание конфигурационного файла приложения по данным общей конфигурации приложения.
func (ccf *impl) Create() (err error) {

	// TODO: Сделать реализацию команды создания конфигурационного файла приложения из структуры конфигурации.

	err = fmt.Errorf(tplCommandNotImplemented, strings.Join([]string{cmdConfig, cmdCreate}, " "))

	return
}
