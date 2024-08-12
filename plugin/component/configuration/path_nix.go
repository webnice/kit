//go:build !windows

package configuration

import "path"

// Функция вернёт список путей, в порядке приоритета, по которым необходимо искать файл конфигурации
func getPath(appName, dirHome, dirWork string, other ...string) (ret []string) {
	const (
		pEtc       = `/etc`
		pOpt       = `/opt`
		pLocEtc    = `/usr/local/etc`
		pShare     = `/usr/share`
		pDotConfig = `.config`
		pConf      = `conf`
	)
	var n int

	// Путь: /etc/[название приложения]/
	ret = append(ret, path.Join(pEtc, appName))
	// Путь: /opt/[название приложения]/.config/
	ret = append(ret, path.Join(pOpt, appName, pDotConfig))
	// Путь: /opt/[название приложения]/conf/
	ret = append(ret, path.Join(pOpt, appName, pConf))
	// Путь: /opt/[название приложения]/
	ret = append(ret, path.Join(pOpt, appName))
	// Путь: /usr/local/etc/[название приложения]/
	ret = append(ret, path.Join(pLocEtc, appName))
	// Путь: /usr/local/etc/[название приложения]/.config/
	ret = append(ret, path.Join(pLocEtc, appName, pDotConfig))
	// Путь: /usr/local/etc/[название приложения]/conf/
	ret = append(ret, path.Join(pLocEtc, appName, pConf))
	// Путь: /usr/share/[название приложения]/.config/
	ret = append(ret, path.Join(pShare, appName, pDotConfig))
	// Путь: /usr/share/[название приложения]/conf/
	ret = append(ret, path.Join(pShare, appName, pConf))
	// Путь: /usr/share/[название приложения]/
	ret = append(ret, path.Join(pShare, appName))
	if dirHome != "" {
		// Путь: [домашняя папка]/.config/[название приложения]/
		ret = append(ret, path.Join(dirHome, pDotConfig, appName))
		// Путь: [домашняя папка]/.config/[название приложения]/conf/
		ret = append(ret, path.Join(dirHome, pDotConfig, appName, pConf))
		// Путь: [домашняя папка]/.config/
		ret = append(ret, path.Join(dirHome, pDotConfig))
		// Путь: [домашняя папка]/conf/
		ret = append(ret, path.Join(dirHome, pConf))
		// Путь: [домашняя папка]/
		ret = append(ret, dirHome)
	}
	if dirWork != "" {
		// Путь: [текущая папка]/.config/
		ret = append(ret, path.Join(dirWork, pDotConfig))
		// Путь: [текущая папка]/conf/
		ret = append(ret, path.Join(dirWork, pConf))
		// Путь: [текущая папка]/
		ret = append(ret, dirWork)
	}
	for n = range other {
		// Путь: other[n]/.config/
		ret = append(ret, path.Join(other[n], pDotConfig))
		// Путь: other[n]/conf/
		ret = append(ret, path.Join(other[n], pConf))
		// Путь: other[n]/
		ret = append(ret, other[n])
	}

	return
}
