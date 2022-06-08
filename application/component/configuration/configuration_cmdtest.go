// Package configuration
package configuration

import (
	"fmt"
	"strings"
)

// Test Тестирование конфигурации приложения.
func (ccf *impl) Test() (err error) {

	// TODO: Сделать реализацию команды тестирования конфигурационного файла приложения.

	err = fmt.Errorf(tplCommandNotImplemented, strings.Join([]string{cmdConfig, cmdTest}, " "))

	return
}
