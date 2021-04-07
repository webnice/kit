package file

import (
	"fmt"
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
	var (
		fi []os.FileInfo
		n  int
	)

	if fi, err = ioutil.ReadDir(pt); err != nil {
		err = fmt.Errorf("read dir error: %s", err)
		return
	}
	if len(fi) == 0 && level > 0 {
		if err = os.Remove(pt); err != nil {
			err = fmt.Errorf("remove %q, error: %s", pt, err)
		}
		return
	}
	for n = range fi {
		switch {
		case fi[n].IsDir():
			if err = fl.CleanEmptyFolderRecursive(path.Join(pt, fi[n].Name()), level+1); err != nil {
				return
			}
		}
	}

	return
}
