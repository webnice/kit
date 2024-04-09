package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// CleanEmptyFolder Удаление пустых папок.
func (fl *impl) CleanEmptyFolder(folderPath string) (err error) {
	return fl.cleanEmptyFolderRecursive(folderPath, 0)
}

// Более удобный для рекурсии вариант функции удаления пустых папок.
func (fl *impl) cleanEmptyFolderRecursive(pt string, level int64) (err error) {
	var (
		fi []os.FileInfo
		n  int
	)

	if fi, err = ioutil.ReadDir(pt); err != nil {
		err = fmt.Errorf("чтение директории %q прервано ошибкой: %s", pt, err)
		return
	}
	if len(fi) == 0 && level > 0 {
		if err = os.Remove(pt); err != nil {
			err = fmt.Errorf("удаление директории %q прервано ошибкой: %s", pt, err)
		}
		return
	}
	for n = range fi {
		switch {
		case fi[n].IsDir():
			if err = fl.cleanEmptyFolderRecursive(path.Join(pt, fi[n].Name()), level+1); err != nil {
				return
			}
		}
	}

	return
}
