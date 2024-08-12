package main

/*

Создание импорта автоматического подключения компонентов, модулей, моделей и иных частей приложения.
В качестве аргументов принимаются пути, относительно корневой директории проекта.
Если в конце пути указано "/...", тогда поиск директорий пакетов производится рекурсивно, в поддиректориях.
Пример:
go:generate go run "github.com/webnice/kit/v4/app/component" "app/component" "ctl/..."

Импорт создаётся со списком найденных пакетов, которые подключаются с использованием синонима "self",
синоним проекта необходимо прописать в go.mod.

*/

import (
	"bytes"
	"go/format"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

const outComponentsGenerated = "components_generated.go" // Имя файла для записи результата.
const (
	defaultProjectAlias   = "self"    // Синоним проекта, определяемый в go.mod.
	defaultIgnoreFilename = ".ignore" // Имя файла, наличие которого, пропускает включение директории в импорт.
)

// Описание просматриваемых директорий.
type tDir struct {
	Absolute   string // Абсолютный путь.
	Relative   string // Относительный путь.
	RasProject string // Путь импорта относительно проекта.
	Recursive  bool   // Просматривать рекурсивно поддиректории.
}

func main() {
	const defaultComponentPath = "."
	var (
		err  error
		cp   []string
		code int
	)

	if len(os.Args) >= 2 {
		cp = append(cp, os.Args[1:]...)
	}
	if len(cp) == 0 {
		cp = append(cp, defaultComponentPath)
	}
	if code, err = makeImports(findRootPath(), cp...); err != nil {
		log.Println(err)
		os.Exit(code)
	}
}

func makeImports(rootPath string, paths ...string) (ret int, err error) {
	const templateBody = `// nolint: lll

// НЕ РЕДАКТИРОВАТЬ! Изменения будут перезаписаны при следующей генерации.
// DO NOT EDIT! Code generated by go generate.

package {{ .Package }}

import (
{{- range .Imports }}
{{ printf "_ %q" .RasProject }}
{{- end }}
)
`
	var (
		tpl  *template.Template
		buf  *bytes.Buffer
		dirs []*tDir
		vars map[string]any
		n    int
		b    []byte
	)

	// Получение списка директорий содержащих пакеты импорта.
	dirs = findDirs(rootPath, paths...)
	// Подготовка данных.
	for n = range dirs {
		dirs[n].RasProject = path.Join(defaultProjectAlias, dirs[n].Relative)
	}
	vars = map[string]any{
		"Package": "main",
		"Imports": dirs,
	}
	tpl = template.Must(
		template.New("").
			Parse(templateBody),
	)
	// Формирование файла импорта.
	buf = &bytes.Buffer{}
	if err = tpl.Execute(buf, vars); err != nil {
		return
	}
	// Форматирование файла импорта с использованием go fmt.
	if b, err = format.Source(buf.Bytes()); err != nil {
		return
	}
	// Запись результата.
	if err = os.WriteFile(outComponentsGenerated, b, 0644); err != nil {
		return
	}

	return
}

// Поиск директорий содержащих go файлы и не содержащих файл-флаг исключения.
func findDirs(rootPath string, paths ...string) (ret []*tDir) {
	const suffixRecursive, extGo = "/...", ".go"
	var (
		err  error
		dirs []*tDir
		dir  *tDir
		n, d int
		dry  []os.DirEntry
	)

	dirs = make([]*tDir, 0, len(paths))
	for n = range paths {
		dir = &tDir{
			Absolute:  path.Clean(path.Join(rootPath, strings.TrimSuffix(paths[n], suffixRecursive))),
			Relative:  path.Clean(strings.TrimSuffix(paths[n], suffixRecursive)),
			Recursive: strings.HasSuffix(paths[n], suffixRecursive),
		}
		dirs = append(dirs, dir)
	}
	for d = range dirs {
		// Сама директория.
		if isPath(dirs[d].Absolute, []string{extGo}, []string{defaultIgnoreFilename}) {
			if ret = append(ret, dirs[d]); !dirs[d].Recursive {
				continue
			}
		}
		// Содержимое директории.
		if dry, err = os.ReadDir(dirs[d].Absolute); err != nil {
			log.Printf("чтение директории %q прервано ошибкой: %s", dirs[d].Absolute, err)
			continue
		}
		for n = range dry {
			if !dry[n].IsDir() {
				continue
			}
			dir = &tDir{
				Absolute:  path.Clean(path.Join(dirs[d].Absolute, dry[n].Name())),
				Relative:  path.Clean(path.Join(dirs[d].Relative, dry[n].Name())),
				Recursive: dirs[d].Recursive,
			}
			if isPath(dir.Absolute, []string{extGo}, []string{defaultIgnoreFilename}) {
				ret = append(ret, dir)
				continue
			}
			if dirs[d].Recursive {
				ret = append(ret, findDirs(rootPath, path.Join(dir.Relative, suffixRecursive))...)
			}
		}
	}

	return
}

// Поиск корневой директории проекта по файлам, которые обязательно должны быть в проекте.
func findRootPath() (ret string) {
	const limit, up, fn1, fn2, fn3 = 4, "..", "Makefile", ".prj.mk", "go.mod"
	var (
		err   error
		fnd   []string
		tmp   string
		try   int
		found bool
	)

	fnd = []string{fn1, fn2, fn3}
	if tmp, err = filepath.Abs(""); err != nil {
		log.Fatalf("Не удалось получить имя выполняемого файла, ошибка: %s", err)
		return
	}
	ret = path.Dir(tmp)
	for ; try <= limit; try++ {
		if found = isPath(ret, fnd, []string{}); found {
			break
		}
		ret = path.Clean(path.Join(ret, up))
	}
	if !found {
		ret = ""
	}

	return
}

// Возвращается истина, если одновременно выполняются следующие условия:
// * указанный путь является директорией;
// * указанный путь содержит файлы, перечисленные в yes;
// * указанный путь не содержит файлы, перечисленные в not;
func isPath(p string, yes []string, not []string) (ret bool) {
	var (
		err  error
		dry  []os.DirEntry
		n, j int
		yesN int
		notN int
	)

	if dry, err = os.ReadDir(p); err != nil {
		return
	}
	for n = range dry {
		for j = range yes {
			if !dry[n].IsDir() && strings.HasSuffix(dry[n].Name(), yes[j]) {
				yesN++
			}
		}
		for j = range not {
			if !dry[n].IsDir() && dry[n].Name() == not[j] {
				notN++
			}
		}
	}
	if yesN > 0 {
		ret = true
	}
	if notN > 0 || yesN < len(yes) {
		ret = false
	}

	return
}
