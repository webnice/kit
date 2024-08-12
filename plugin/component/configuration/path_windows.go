//go:build windows

package configuration

import (
	"path"
	"strings"
	"syscall"
)

// Функция вернёт список путей, в порядке приоритета, по которым необходимо искать файл конфигурации
func getPath(appName, dirHome, dirWork string, other ...string) (ret []string) {
	const (
		pDotConfig          = `.config`
		pConf               = `conf`
		eSYSTEMROOT         = "SYSTEMROOT"
		ePROGRAMFILES       = "PROGRAMFILES"
		eCOMMONPROGRAMFILES = "COMMONPROGRAMFILES"
		ePROGRAMDATA        = "PROGRAMDATA"
		eLOCALAPPDATA       = "LOCALAPPDATA"
	)
	var (
		n       int
		envName string
		tmp     string
	)

	for _, envName = range []string{ePROGRAMDATA, ePROGRAMFILES, eCOMMONPROGRAMFILES, eSYSTEMROOT, eLOCALAPPDATA} {
		envName = getEnv(envName)
		// Путь: SYSTEM[n]/[название приложения]/.config/
		ret = append(ret, path.Join(envName, appName, pDotConfig))
		// Путь: SYSTEM[n]/[название приложения]/conf/
		ret = append(ret, path.Join(envName, appName, pConf))
		// Путь: SYSTEM[n]/[название приложения]/
		ret = append(ret, path.Join(envName, appName))
	}
	if dirHome != "" {
		tmp = strings.Replace(dirHome, symbolBackslash, symbolSlash, -1)
		// Путь: [домашняя папка]/.config/[название приложения]/
		ret = append(ret, path.Join(tmp, pDotConfig, appName))
		// Путь: [домашняя папка]/.config/[название приложения]/conf/
		ret = append(ret, path.Join(tmp, pDotConfig, appName, pConf))
		// Путь: [домашняя папка]/.config/
		ret = append(ret, path.Join(tmp, pDotConfig))
		// Путь: [домашняя папка]/conf/
		ret = append(ret, path.Join(tmp, pConf))
		// Путь: [домашняя папка]/
		ret = append(ret, tmp)
	}
	if dirWork != "" {
		tmp = strings.Replace(dirWork, symbolBackslash, symbolSlash, -1)
		// Путь: [текущая папка]/.config/
		ret = append(ret, path.Join(tmp, pDotConfig))
		// Путь: [текущая папка]/conf/
		ret = append(ret, path.Join(tmp, pConf))
		// Путь: [текущая папка]/
		ret = append(ret, tmp)
	}
	for n = range other {
		tmp = strings.Replace(other[n], symbolBackslash, symbolSlash, -1)
		// Путь: other[n]/.config/
		ret = append(ret, path.Join(tmp, pDotConfig))
		// Путь: other[n]/conf/
		ret = append(ret, path.Join(tmp, pConf))
		// Путь: other[n]/
		ret = append(ret, tmp)
	}

	return
}

func getEnv(envName string) (ret string) {
	var ok bool

	if ret, ok = syscall.Getenv(envName); ok {
		ret = strings.Replace(ret, symbolBackslash, symbolSlash, -1)
	}

	return
}
