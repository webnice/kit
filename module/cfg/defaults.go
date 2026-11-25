package cfg

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	kitModuleCfgCli "github.com/webnice/kit/v5/module/cfg/cli"
	kitModuleCfgConst "github.com/webnice/kit/v5/module/cfg/const"
	kitModuleLogLevel "github.com/webnice/kit/v5/module/log/level"
)

// Целевой уровень выполнения приложения.
func defaultApplicationTargetlevelUint16() (ret uint16) {
	const defaultApplicationTargetlevel = 65535

	ret = defaultApplicationTargetlevel

	return
}

// Целевой уровень выполнения приложения.
func defaultApplicationTargetlevelString() string {
	return strconv.FormatUint(uint64(defaultApplicationTargetlevelUint16()), 10)
}

// Режим отладки приложения.
func defaultApplicationDebugBool() (ret bool) {
	const defaultApplicationDebug = false

	ret = defaultApplicationDebug

	return
}

// Режим отладки приложения.
func defaultApplicationDebugString() string { return strconv.FormatBool(defaultApplicationDebugBool()) }

// Название приложения.
func defaultApplicationName() (ret string) {
	var tmp []string

	if tmp = strings.Split(os.Args[0], string(os.PathSeparator)); len(tmp) > 0 {
		ret = tmp[len(tmp)-1]
	}
	switch runtime.GOOS {
	case keyOsWindows:
		tmp = strings.Split(filepath.Base(ret), `.`)
		ret = tmp[len(tmp)-2]
	}

	return
}

// Описание приложения или сервиса, отображаемое в строке помощи.
func defaultApplicationDescription() (ret string) {
	const tpl = "--- ### %s ### ---"

	ret = fmt.Sprintf(tpl, defaultApplicationName())

	return
}

// Домашняя директория пользователя, от имени которого выполняется приложение.
func defaultHomeDirectory() (ret string) {
	const homePathSymbol = "~"
	var err error

	if ret, err = os.UserHomeDir(); err != nil || ret == "" {
		ret = homePathSymbol
	}

	return
}

// Рабочая директория приложения.
func defaultWorkingDirectory() (ret string) {
	const defaultRootPath = "/"
	var err error

	if defaultWorkingDirectoryOriginal == "" {
		if defaultWorkingDirectoryOriginal, err = os.Getwd(); err != nil {
			defaultWorkingDirectoryOriginal = defaultRootPath
		}
		defaultWorkingDirectoryOriginal = AbsolutePath(defaultWorkingDirectoryOriginal)
	}
	ret = defaultWorkingDirectoryOriginal

	return
}

// Директория для временных файлов.
func defaultTempDirectory() (ret string) {
	var (
		tmp string
		err error
	)

	if defaultTempDirectoryOriginal == "" {
		defaultTempDirectoryOriginal = os.TempDir()
		if tmp, err = filepath.EvalSymlinks(defaultTempDirectoryOriginal); err == nil {
			defaultTempDirectoryOriginal = tmp
		}
		defaultTempDirectoryOriginal = path.Clean(defaultTempDirectoryOriginal)
		defaultTempDirectoryOriginal = AbsolutePath(defaultTempDirectoryOriginal)
	}
	ret = defaultTempDirectoryOriginal

	return
}

// Директория для файлов кеширования.
func defaultCacheDirectory() (ret string) {
	const (
		homePathSymbol = "~"
		unixPath       = "/var/cache"
		windowsPath    = "AppData\\Local\\Packages"
	)
	var err error

	if defaultCacheDirectoryOriginal == "" {
		if defaultCacheDirectoryOriginal, err = os.UserCacheDir(); err != nil {
			switch runtime.GOOS {
			case keyOsWindows:
				defaultCacheDirectoryOriginal = path.Join(homePathSymbol, windowsPath, defaultApplicationName())
			default:
				defaultCacheDirectoryOriginal = path.Join(unixPath, defaultApplicationName())
			}
		}
		defaultCacheDirectoryOriginal = AbsolutePath(defaultCacheDirectoryOriginal)
	}
	ret = defaultCacheDirectoryOriginal

	return
}

// Директория конфигурации в домашней директории пользователя.
func defaultConfigDirectory() (ret string) {
	const (
		homePathSymbol = "~"
		unixPath       = ".config"
		windowsPath    = "AppData\\Roaming"
	)
	var err error

	if defaultConfigDirectoryOriginal == "" {
		if defaultConfigDirectoryOriginal, err = os.UserConfigDir(); err != nil {
			switch runtime.GOOS {
			case keyOsWindows:
				defaultConfigDirectoryOriginal = path.Join(homePathSymbol, windowsPath, defaultApplicationName())
			default:
				defaultConfigDirectoryOriginal = path.Join(`~`, unixPath, defaultApplicationName())
			}
		}
		defaultConfigDirectoryOriginal = AbsolutePath(defaultConfigDirectoryOriginal)
	}
	ret = defaultConfigDirectoryOriginal

	return
}

// Директория для файлов журнала приложения.
func defaultLogDirectory() (ret string) {
	const (
		homePathSymbol = "~"
		unixPath       = "/var/log"
		windowsPath    = "AppData\\Local"
	)
	var err error

	if defaultLogDirectoryOriginal == "" {
		if defaultLogDirectoryOriginal, err = os.UserHomeDir(); err != nil {
			switch runtime.GOOS {
			case keyOsWindows:
				defaultLogDirectoryOriginal = path.Join(homePathSymbol, windowsPath, defaultApplicationName())
			default:
				defaultLogDirectoryOriginal = path.Join(unixPath, defaultApplicationName())
			}
		}
		defaultLogDirectoryOriginal = AbsolutePath(defaultLogDirectoryOriginal)
	}
	ret = defaultLogDirectoryOriginal

	return
}

// Время отводимое на инициализацию одного компонента.
func defaultInitiateTimeout() (ret time.Duration) {
	const defaultMinutes = 1

	ret = time.Minute * defaultMinutes

	return
}

// Время отводимое на выполнение функции Finalize() до печати в лог warning сообщения.
func defaultFinalizeTimeoutBeforeWarning() (ret time.Duration) {
	const defaultSecond = 30

	ret = time.Second * defaultSecond

	return
}

// Уровень логирования по умолчанию до загрузки конфигурации.
func defaultLogLevel() (ret kitModuleLogLevel.Level) {
	ret = kitModuleLogLevel.Off

	return
}

// Имена переменных окружения приложения по умолчанию.
func defaultConstantEnvironment() (ret kitModuleCfgCli.ConstantEnvironmentName) {
	ret = kitModuleCfgCli.ConstantEnvironmentName{
		{Anchor: kitModuleCfgConst.EnvironmentApplicationForkWorkerMaster, Destination: envNameAnchorForApplicationForkWorkerMaster},
		{Anchor: kitModuleCfgConst.EnvironmentApplicationForkWorkerComponent, Destination: envNameAnchorForApplicationForkWorkerComponent},
		{Anchor: kitModuleCfgConst.EnvironmentApplicationForkWorkerTarget, Destination: envNameAnchorForApplicationForkWorkerTarget},
		{Anchor: kitModuleCfgConst.EnvironmentApplicationTargetlevel, Destination: envNameAnchorForApplicationTargetlevel},
		{Anchor: kitModuleCfgConst.EnvironmentApplicationDebug, Destination: envNameAnchorForApplicationDebug},
		{Anchor: kitModuleCfgConst.EnvironmentApplicationName, Destination: envNameAnchorForApplicationName},
		{Anchor: kitModuleCfgConst.EnvironmentHomeDirectory, Destination: envNameAnchorForHomeDirectory},
		{Anchor: kitModuleCfgConst.EnvironmentWorkingDirectory, Destination: envNameAnchorForWorkingDirectory},
		{Anchor: kitModuleCfgConst.EnvironmentTempDirectory, Destination: envNameAnchorForTempDirectory},
		{Anchor: kitModuleCfgConst.EnvironmentCacheDirectory, Destination: envNameAnchorForCacheDirectory},
		{Anchor: kitModuleCfgConst.EnvironmentConfigDirectory, Destination: envNameAnchorForConfigDirectory},
		{Anchor: kitModuleCfgConst.EnvironmentLogDirectory, Destination: envNameAnchorForLogDirectory},
		{Anchor: kitModuleCfgConst.EnvironmentConfigFile, Destination: envNameAnchorForConfigFile},
		{Anchor: kitModuleCfgConst.EnvironmentPidFile, Destination: envNameAnchorForPidFile},
		{Anchor: kitModuleCfgConst.EnvironmentStateFile, Destination: envNameAnchorForStateFile},
		{Anchor: kitModuleCfgConst.EnvironmentSocketFile, Destination: envNameAnchorForSocketFile},
		{Anchor: kitModuleCfgConst.EnvironmentLogLevel, Destination: envNameAnchorForLogLevel},
		{Anchor: kitModuleCfgConst.EnvironmentLogGraylog, Destination: envNameAnchorForLogGraylog},
	}

	return
}
