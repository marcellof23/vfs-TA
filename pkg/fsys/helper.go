package fsys

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcellof23/vfs-TA/boot"
)

type WalkDirFunc func(path string, fs *filesystem, err error) error

func New() *filesystem {
	// uncomment for recursively grab all files and directories from this level downwards.
	root = replicateFilesystem(".", nil)

	// root = makeFilesystem(".", ".", nil)
	fsys := root
	return fsys
}

// testFilessytemCreation initializes the filesystem by replicating
// the current root directory and all it's child direcctories.
func replicateFilesystem(dirName string, fs *filesystem) *filesystem {
	var fi os.FileInfo
	var fileName os.FileInfo

	if dirName == "." {
		root = makeFilesystem(".", ".", nil)
		fs = root
	}
	index := 0
	files, _ := ioutil.ReadDir(dirName)
	for index < len(files) {
		fileName = files[index]
		fi, _ = os.Stat(dirName + "//" + fileName.Name())
		mode := fi.Mode()

		if mode.IsDir() {
			if fileName.Name() != "vendor" && fileName.Name() != ".git" {
				fs.directories[fileName.Name()] = makeFilesystem(fileName.Name(), strings.ReplaceAll(dirName, "//", "/")+"/"+fileName.Name(), fs)
				err := fs.MFS.MkdirAll(fileName.Name(), 0o700)
				if err != nil {
					fmt.Println(err)
				}
				replicateFilesystem(dirName+"//"+fileName.Name(), fs.directories[fileName.Name()])
			}

		} else {
			if fileName.Name() != "vendor" && fileName.Name() != ".git" {
				fs.files[fileName.Name()] = &file{
					name:     fileName.Name(),
					rootPath: strings.ReplaceAll(dirName, "//", "/") + "/" + fileName.Name(),
				}

				fs.MFS.Create(fs.files[fileName.Name()].rootPath)
			}
		}

		index++
	}
	return fs
}

func makeFilesystem(dirName string, rootPath string, prev *filesystem) *filesystem {
	fs := boot.InitFilesystem()
	return &filesystem{
		fs,
		&fileDir{
			name:        dirName,
			rootPath:    rootPath,
			files:       make(map[string]*file),
			directories: make(map[string]*filesystem),
			prev:        prev,
		},
	}
}

func walkDir(fsys *filesystem, path string, walkDirFn WalkDirFunc) error {

	err := walkDirFn(path, fsys, nil)
	if err != nil {
		return err
	}

	if fsys.files != nil {
		for _, fl := range fsys.files {
			var pathName string
			if path != "." {
				pathName = filepath.Join(path, fl.name)
			} else {
				pathName = fl.name
			}
			walkDirFn(pathName, fsys, nil)
		}
	}

	if len(fsys.directories) > 0 {
		for dirName := range fsys.directories {
			name1 := filepath.Join(fsys.rootPath, dirName)
			if err := walkDir(fsys.directories[dirName], name1, walkDirFn); err != nil {
				return err
			}
		}
	}

	return nil
}

func Walkdir(fsys *filesystem, pathTarget string, walkDirFn WalkDirFunc) error {
	if fsys.rootPath == pathTarget {
		err := walkDir(fsys, pathTarget, walkDirFn)
		if err != nil {
			return err
		}
	}

	if fsys.files != nil {
		for _, _ = range fsys.files {
			err := walkDir(fsys, pathTarget, walkDirFn)
			if err != nil {
				return err
			}
		}
	}

	if len(fsys.directories) > 0 {
		for dirName := range fsys.directories {
			fmt.Println("Directories: ", dirName)
			name1 := filepath.Join(fsys.rootPath, dirName)
			if err := walkDir(fsys.directories[dirName], name1, walkDirFn); err != nil {
				return err
			}
		}
	}

	return nil
}
