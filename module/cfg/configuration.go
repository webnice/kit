// Package cfg
package cfg

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	runtimeDebug "runtime/debug"

	kitModuleCfgCpy "github.com/webnice/kit/module/cfg/cpy"
	kitTypes "github.com/webnice/kit/types"

	"gopkg.in/yaml.v3"
)

// ConfigurationRegistration Регистрация объекта конфигурации, являющегося частью общей конфигурации приложения.
// Регистрация доступна только на уровне работы приложения 0.
// Объект конфигурации должен передаваться в качестве адреса.
// Поля объекта конфигурации должны состоять из простых или сериализуемых типов данных и должны быть экспортируемыми.
// Поля объекта должны содержать теги:
// * description ---- Описание поля, публикуется в YAML файле, при создании примера конфигурации,
//                    подробности в компоненте "configuration".
//                    Если указано значение "-", тогда описание не публикуется.
// * env-name ------- Название переменной конфигурации, из которой загружается значение, если значение в конфигурации
//                    отсутствует или является пустым значением.
// * default-value -- Значение переменной по умолчанию, присваивается после чтения конфигурационного файла,
//                    а так же, публикуется в YAML файле, при создании примера конфигурации.
// * yaml ----------- Тег для библиотеки YAML, если указано значение "-", тогда поле пропускается.
func (essence *gist) ConfigurationRegistration(c interface{}) (err error) {
	const (
		tplDupName = "объект с типом %T, содержит поле с именем %q, которое совпадает с именем поля другого объекта" +
			" конфигурации"
		tplDupTag = "объект с типом %T, содержит поле с тегом %q, со значением %q, которое совпадает со значением" +
			" тега %q, из поля другого объекта конфигурации"
	)
	var (
		item  *configurationItem
		n     int
		tsf   reflect.StructField
		val   reflect.Value
		value string
		ok    bool
	)

	defer func() {
		if e := recover(); e != nil {
			err = Errors().ConfigurationApplicationPanic(0, e, runtimeDebug.Stack())
		}
	}()
	// Проверка уровня работы приложения.
	if essence.parent.runLevel > 0 {
		err = essence.parent.Errors().ConfigurationApplicationProhibited(0, reflect.TypeOf(c).String())
		return
	}
	// Создание объекта конфигурации. Проверка корректности объекта структуры конфигурации.
	item = &configurationItem{Original: c}
	if item.Value, item.Type, err = reflectStructObject(c); err != nil {
		return
	}
	// Создание среза длинной равной количеству полей структуры конфигурации.
	item.Fields = make([]reflect.StructField, 0, item.Type.NumField())
	// Обход всех полей с проверкой и добавлением в срез.
	for n = 0; n < item.Type.NumField(); n++ {
		tsf, val = item.Type.Field(n), item.Value.Field(n)
		// Фильтрация по свойствам поля
		if !val.CanSet() || !tsf.IsExported() || tsf.Anonymous {
			continue
		}
		// Проверка наличия поля с тем же именем в уже добавленных структурах конфигураций.
		if essence.parent.conf.IsName(tsf.Name) {
			err = fmt.Errorf(tplDupName, c, tsf.Name)
			return
		}
		// Проверка тега yaml.
		if value, ok = tsf.Tag.Lookup(tagYaml); ok {
			// Фильтрация по значению тега поля, пропуск полей с yaml:"-".
			if value == "-" {
				continue
			}
			// Проверка наличия тега с тем же синонимом в уже добавленных полях всех добавленных объектов конфигураций,
			// иначе, при загрузке YAML файла, библиотека yaml упадёт в панику, на повторяющемся теге.
			if essence.parent.conf.IsTagValue(tagYaml, value) {
				err = fmt.Errorf(tplDupTag, c, tagYaml, value, tagYaml)
				return
			}
		}
		// Удаление всех тегов, кроме используемых, лишние данные в памяти не нужны.
		tsf.Tag = reflectCleanStructTag(tsf.Tag, tagYaml, tagDefaultValue, tagEnvName, tagDescription)
		// Добавление в результирующий массив.
		item.Fields = append(item.Fields, tsf)
	}
	essence.parent.conf.Items = append(essence.parent.conf.Items, item)

	return
}

// Копирование части параметров стартовой конфигурации приложения из загруженного файла конфигурации.
func (essence *gist) configurationCopyBootstrapConfiguration() {
	var (
		sc = essence.parent.loadableConfiguration
		dc = essence.parent.bootstrapConfiguration
	)

	// Основное правило копирования значений переменных:
	// Если переменная равна значению по умолчанию и если загруженное значение переменной не равно значению
	// по умолчанию golang, тогда загруженное значение присваивается конфигурации.
	// Результатом будет установка значения в порядке убывания приоритета:
	// 1. Значение переданное в командной строке.
	// 2. Значение переданное в переменной окружения "env" блока "kong".
	// 3. Значение по умолчанию указанное в блоке "kong".
	// 4. Значение загруженное из конфигурационного файла.
	// 5. Значение указанное в переменной окружения "env-name".
	// 6. Значение по умолчанию указанное в "default-value".
	// 7. Значение типа переменной golang (пустое значение).
	//
	// Целевой уровень выполнения приложения.
	if dc.ApplicationTargetlevel == defaultApplicationTargetlevelUint16() && sc.ApplicationTargetlevel != 0 {
		dc.ApplicationTargetlevel = sc.ApplicationTargetlevel
	}
	// Включение режима отладки приложения.
	if dc.ApplicationDebug == defaultApplicationDebugBool() && sc.ApplicationDebug != false {
		dc.ApplicationDebug = sc.ApplicationDebug
	}
	// Название приложения.
	if dc.ApplicationName == defaultApplicationName() && sc.ApplicationName != "" {
		dc.ApplicationName = sc.ApplicationName
	}
	// Домашняя директория приложения.
	if dc.HomeDirectory == defaultHomeDirectory() && sc.HomeDirectory != "" {
		dc.HomeDirectory = essence.parent.AbsolutePath(sc.HomeDirectory)
	}
	// Рабочая директория приложения.
	if dc.WorkingDirectory == defaultWorkingDirectory() && sc.WorkingDirectory != "" {
		dc.WorkingDirectory = essence.parent.AbsolutePath(sc.WorkingDirectory)
	}
	// Директория для временных файлов.
	if dc.TempDirectory == defaultTempDirectory() && sc.TempDirectory != "" {
		dc.TempDirectory = essence.parent.AbsolutePath(sc.TempDirectory)
	}
	// Директория для файлов кеша.
	if dc.CacheDirectory == defaultCacheDirectory() && sc.CacheDirectory != "" {
		dc.CacheDirectory = essence.parent.AbsolutePath(sc.CacheDirectory)
	}
	// Директория для подключаемых или дополнительных конфигураций приложения.
	if dc.ConfigDirectory == defaultConfigDirectory() && sc.ConfigDirectory != "" {
		dc.ConfigDirectory = essence.parent.AbsolutePath(sc.ConfigDirectory)
	}
	// Путь и имя PID файла приложения.
	if dc.PidFile == "" && sc.PidFile != "" {
		dc.PidFile = essence.parent.AbsolutePath(sc.PidFile)
	}
	// Путь и имя файла хранения состояния приложения.
	if dc.StateFile == "" && sc.StateFile != "" {
		dc.StateFile = essence.parent.AbsolutePath(sc.StateFile)
	}
	// Сокет файл коммуникаций с приложением, только для *nix систем, путь и имя файла.
	if dc.SocketFile == "" && sc.SocketFile != "" {
		dc.SocketFile = essence.parent.AbsolutePath(sc.SocketFile)
	}
	// Уровень логирования по умолчанию до загрузки конфигурации приложения.
	if dc.LogLevel == defaultLogLevel() && sc.LogLevel != 0 {
		dc.LogLevel = sc.LogLevel
	}
}

// Установка значений по умолчанию, значениями из переменных окружения, значениями описанными в тегах в структуры или
// значениями полученными из метода Default(), если структура реализует интерфейс types.ConfigurationDefaulter.
// Поддерживаются все простые типы данных, а так же типы данных, у которых реализованы
// интерфейсы sql.Scanner и encoding.TextUnmarshaler
func (essence *gist) configurationSetDefaultValue() (err error) {
	var csdv func(dc interface{})

	defer func() {
		if e := recover(); e != nil {
			err = Errors().ConfigurationSetDefaultPanic(0, e, runtimeDebug.Stack())
		}
	}()
	csdv = func(dc interface{}) {
		var (
			dcRt           reflect.Type
			dcRv           reflect.Value
			dcRsf          reflect.StructField
			dcV            reflect.Value
			n              int
			ok, found      bool
			envName        string
			defaultValue   string
			scanner        sql.Scanner
			scannerRv      reflect.Value
			defaultValueRv reflect.Value
			tcdi           kitTypes.ConfigurationDefaulter
			objDefaultRv   reflect.Value
		)

		if dcRv, dcRt, err = reflectStructObject(dc); err != nil {
			err = essence.parent.Errors().ConfigurationSetDefault(0, err)
			return
		}
		for n = 0; n < dcRt.NumField(); n++ {
			dcRsf, dcV = dcRt.Field(n), dcRv.Field(n)
			if !dcV.CanSet() || !dcRsf.IsExported() || dcRsf.Anonymous {
				continue
			}
			if dcV.IsZero() {
				found, ok = false, false
				if envName, ok = dcRsf.Tag.Lookup(tagEnvName); ok && envName != "-" {
					if defaultValue, ok = os.LookupEnv(envName); ok && defaultValue != "" {
						found = true
					}
				}
				if !found {
					defaultValue, ok = dcRsf.Tag.Lookup(tagDefaultValue)
					if ok && defaultValue != "-" && defaultValue != "" {
						found = true
					}
				}
				if !found {
					continue
				}
				scanner = makeScanner(dcV)
				if scannerRv, _, err = reflectObject(scanner); err != nil {
					err = essence.parent.Errors().ConfigurationSetDefault(0, err)
					return
				}
				if defaultValueRv, _, err = reflectObject(&defaultValue); err != nil {
					err = essence.parent.Errors().ConfigurationSetDefault(0, err)
					return
				}
				if ok, err = kitModuleCfgCpy.Gist().Set(scannerRv, defaultValueRv); err != nil {
					err = essence.parent.Errors().ConfigurationSetDefaultValue(0, defaultValue, dcRsf.Name, err)
					return
				}
			} else if dcV.CanAddr() && dcRsf.Type.Kind() == reflect.Struct {
				// Проверка на реализацию интерфейса types.ConfigurationDefaulter
				switch tcdi, ok = dcV.Addr().Interface().(kitTypes.ConfigurationDefaulter); ok {
				case true:
					objDefaultRv = reflect.New(dcRsf.Type)
					if tcdi, ok = objDefaultRv.Interface().(kitTypes.ConfigurationDefaulter); ok {
						if err = tcdi.Default(); err != nil {
							err = essence.parent.Errors().ConfigurationSetDefault(0, err)
							return
						}
						if err = kitModuleCfgCpy.Gist().CopyToIsZero(dcV, objDefaultRv); err != nil {
							err = essence.parent.Errors().ConfigurationSetDefault(0, err)
							return
						}
					}
				default:
					csdv(dcV.Addr().Interface())
				}
			}
		}
	}
	// Union всегда является типом interface{} и не может реализовывать интерфейс types.ConfigurationDefaulter, поэтому
	// данный вызов не нуждается в проверке на реализацию интерфейса types.ConfigurationDefaulter.
	csdv(essence.parent.conf.Union)

	return
}

// ConfigurationLoad Загрузка конфигурационного файла
func (essence *gist) ConfigurationLoad(buf *bytes.Buffer) (err error) {
	const tplYamlDecodeError = "декодирование конфигурации из формата yaml прервано ошибкой: %s"
	var (
		cObjectStructType reflect.Type
		yamlDecoder       *yaml.Decoder
		n                 int
	)

	defer func() {
		if e := recover(); e != nil {
			err = Errors().ConfigurationApplicationPanic(0, e, runtimeDebug.Stack())
		}
	}()
	// Создание новой структуры конфигурации объединяющей в себе все зарегистрированные структуры конфигураций.
	cObjectStructType = reflect.StructOf(essence.parent.conf.StructField())
	essence.parent.conf.Union = reflect.New(cObjectStructType).Interface()
	yamlDecoder = yaml.NewDecoder(buf)
	if err = yamlDecoder.Decode(essence.parent.conf.Union); err != nil {
		err = fmt.Errorf(tplYamlDecodeError, err)
		return
	}
	// Копирование данных из объединённой конфигурации в зарегистрированные объекты конфигураций.
	// Используется локальная копия моей библиотеки: https://github.com/webnice/cpy/.
	for n = range essence.parent.conf.Items {
		if err = kitModuleCfgCpy.All(essence.parent.conf.Items[n].Original, essence.parent.conf.Union); err != nil {
			return
		}
	}
	// Копирование части параметров стартовой конфигурации приложения из загруженного файла конфигурации в стартовую.
	essence.configurationCopyBootstrapConfiguration()
	// Установка значений по умолчанию, если после загрузки данных из файла конфигурации, значение осталось пустым.
	if err = essence.configurationSetDefaultValue(); err != nil {
		return
	}
	// Актуализация данных в объединённой структуре конфигурации.
	if err = kitModuleCfgCpy.All(essence.parent.conf.Union, essence.parent.bootstrapConfiguration); err != nil {
		return
	}

	return
}
