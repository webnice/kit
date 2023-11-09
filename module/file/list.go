package file

import (
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
		deo []os.DirEntry
		n   int
		tmp []string
	)

	pf = path.Join(baseFolderPath, currentFolderPath)
	if deo, err = os.ReadDir(pf); err != nil {
		return
	}
	for n = range deo {
		switch {
		case deo[n].IsDir():
			tmp, err = fl.recursiveFileListLoop(baseFolderPath, path.Join(currentFolderPath, deo[n].Name()))
			if err != nil {
				return
			}
			ret = append(ret, tmp...)
		case !deo[n].IsDir():
			ret = append(ret, path.Join(currentFolderPath, deo[n].Name()))
		}
	}

	return
}
