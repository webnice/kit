package file

import (
	"io/ioutil"
	"os"
	"path"
)

// CleanEmptyFolder Удаление пустых папок
func (fl *impl) CleanEmptyFolder(pt string) (err error) {
	return fl.CleanEmptyFolderRecursive(pt, 0)
}

// CleanEmptyFolderRecursive Более удобный для рекурсии вариант
func (fl *impl) CleanEmptyFolderRecursive(pt string, level int64) (err error) {
	var fi []os.FileInfo
	fi, err = ioutil.ReadDir(pt)
	if len(fi) == 0 && level > 0 {
		err = os.Remove(pt)
		return
	}
	for i := range fi {
		if fi[i].IsDir() {
			err = fl.CleanEmptyFolderRecursive(path.Join(pt, fi[i].Name()), level+1)
			if err != nil {
				return
			}
		}
	}
	return
}
