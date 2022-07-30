// Package file
package file

import (
	"io/ioutil"
	"os"
	"path"
)

// RecursiveFileList Поиск всех файлов начиная от folderPath рекурсивно.
// Возвращается слайс относительных имён файлов.
func (fl *impl) RecursiveFileList(folderPath string) (ret []string, err error) {
	ret, err = fl.recursiveFileListLoop(folderPath, "")

	return
}

// Удобная для рекурсии функция
func (fl *impl) recursiveFileListLoop(baseFolderPath, currentFolderPath string) (ret []string, err error) {
	var (
		pf  string
		fis []os.FileInfo
		n   int
	)

	pf = path.Join(baseFolderPath, currentFolderPath)
	if fis, err = ioutil.ReadDir(pf); err != nil {
		return
	}
	for n = range fis {
		switch {
		case fis[n].IsDir():
			ret, err = fl.recursiveFileListLoop(baseFolderPath, path.Join(currentFolderPath, fis[n].Name()))
			if err != nil {
				return
			}
		case !fis[n].IsDir():
			ret = append(ret, path.Join(currentFolderPath, fis[n].Name()))
		}
	}

	return
}
