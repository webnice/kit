// Package configuration
package configuration

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/webnice/kit/v3/application/component/configuration/osext"
)

// Поиск пути и имени файла конфигурации.
func (ccf *impl) findConfigurationFile(appName, dirConf, dirHome, dirWork string) (ret string) {
	var (
		err         error
		filename    string
		directories []string
		other       []string
		n           int
		fi          os.FileInfo
	)

	// Получение пути к вышестоящей папке от запущенного файла приложения.
	if filename, err = osext.New().Executable(); err == nil {
		filename = path.Join(filepath.Dir(filename), pathParent)
		other = append(other, ccf.cfg.AbsolutePath(filepath.Clean(filename)))
	}
	// Составление платформо-зависимого списка директорий поиска конфигурационного файла.
	if dirConf != "" {
		switch runtime.GOOS {
		case osWindows:
			dirConf = strings.Replace(dirConf, symbolBackslash, symbolSlash, -1)
		}
		// Приоритет за специальной директорией для конфигурации, если такая указана.
		directories = append(directories, dirConf)
	}
	// Стандартные платформо-зависимые пути
	directories = append(directories, getPath(appName, dirHome, dirWork, other...)...)
	// Если списка путей нет, ничего не ищем.
	if len(directories) == 0 {
		return
	}
	// Порядок директорий в массиве описывает приоритет.
	for n = range directories {
		filename = fmt.Sprintf("%s%s", appName, extensionYaml)
		filename = path.Join(directories[n], filename)
		if fi, err = os.Stat(filename); err != nil {
			continue
		}
		if fi.IsDir() {
			continue
		}
		if ret = filename; ret != "" {
			break
		}
	}

	return
}
