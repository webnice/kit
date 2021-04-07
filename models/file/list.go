package file

import (
	"io/fs"
	"io/ioutil"
	"path"
)

// RecursiveFileList Поиск всех файлов начиная от path рекурсивно
// Возвращается слайс относительных имён файлов
func (fl *impl) RecursiveFileList(path string) (ret []string, err error) {
	ret, err = fl.RecursiveFileListLoop(path, "")

	return
}

// RecursiveFileListLoop Удобная для рекурсии функция
func (fl *impl) RecursiveFileListLoop(base, current string) (ret []string, err error) {
	var (
		pf  string
		fis []fs.FileInfo
		n   int
	)

	pf = path.Join(base, current)
	if fis, err = ioutil.ReadDir(pf); err != nil {
		return
	}
	for n = range fis {
		switch {
		case fis[n].IsDir():
			if ret, err = fl.RecursiveFileListLoop(base, path.Join(current, fis[n].Name())); err != nil {
				return
			}
		case !fis[n].IsDir():
			ret = append(ret, path.Join(current, fis[n].Name()))
		}
	}

	return
}
