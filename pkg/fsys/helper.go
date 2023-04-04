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

// Initiation Filesystem

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

// helper function to check file or dir exists

func (fs *filesystem) doesDirExists(pathName string) bool {
	if pathName[0] == '/' {
		info, err := fs.MFS.Stat("." + pathName)
		if err == nil && info.IsDir() {
			return true
		}
	} else {
		info, err := fs.MFS.Stat(filepath.Join(fs.rootPath, pathName))
		if err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

func (fs *filesystem) doesFileExists(pathName string) bool {
	if pathName[0] == '/' {
		info, err := fs.MFS.Stat("." + pathName)
		if err == nil && !info.IsDir() {
			return true
		}
	} else {
		info, err := fs.MFS.Stat(filepath.Join(fs.rootPath, pathName))
		if err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

func (s *shell) doesDirExist(dirName string, fs *filesystem) bool {
	if _, found := fs.directories[dirName]; found {
		return true
	}
	return false
}

func (s *shell) verifyPath(dirName string) (*filesystem, error) {
	checker := s.handleRootNav(dirName)
	segments := strings.Split(dirName, "/")

	for _, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		if segment == ".." {
			if checker.prev == nil {
				continue
			}
			checker = checker.prev
		} else if s.doesDirExist(segment, checker) == true {
			checker = checker.directories[segment]
		} else {
			fmt.Printf("Error : %s doesn't exist\n", dirName)
			return s.Fs, fmt.Errorf("Error : %s doesn't exist\n", dirName)
		}
	}
	return checker, nil
}

func (s *shell) handleRootNav(dirName string) *filesystem {
	if dirName[0] == '/' {
		return root
	}
	return s.Fs
}

func (s *shell) reassemble(dirPath []string) string {
	counter := 1
	var finishedPath string

	finishedPath = dirPath[0]
	for counter < len(dirPath)-1 {
		finishedPath = finishedPath + "/" + dirPath[counter]
		counter++
	}
	return finishedPath
}

func (s *shell) readFile(filename string) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	fmt.Println(string(dat))
}

func walkDir(fsys *filesystem, path string, walkDirFn WalkDirFunc) error {
	err := walkDirFn(path, fsys, nil)
	if err != nil {
		return err
	}

	if fsys.files != nil {
		for _, fl := range fsys.files {
			//var pathName string
			//if path != "." {
			//	pathName = filepath.Join(path, fl.name)
			//} else {
			//	pathName = fl.name
			//}
			walkDirFn(fl.name, fsys, nil)
		}
	}

	if len(fsys.directories) > 0 {
		for dirName := range fsys.directories {
			name1 := filepath.Join(fsys.rootPath, dirName)
			if err := walkDir(fsys.directories[dirName], name1, walkDirFn); err != nil {
				return err
			}
			walkDirFn(dirName, fsys, nil)
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
