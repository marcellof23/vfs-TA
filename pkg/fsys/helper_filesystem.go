package fsys

import (
	"fmt"
	gofs "io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcellof23/vfs-TA/boot"
)

type WalkDirFunc func(path string, fs *filesystem, err error) error

// Initiation Virtual Filesystem

func New() *filesystem {
	// uncomment for recursively grab all files and directories from this level downwards.
	root = replicateFilesystem(".", nil)

	// uncomment for initiate empty virtual filesystem
	// root = makeFilesystem(".", ".", nil)

	fsys := root
	return fsys
}

// TODO: fix this function, pass MFS

// testFilessytemCreation initializes the filesystem by replicating
// the current root directory and all it's child direcctories.
func replicateFilesystem(dirName string, fs *filesystem) *filesystem {
	var fileName gofs.DirEntry
	var fi os.FileInfo

	if dirName == "." {
		root = makeFilesystem(".", ".", nil, nil)
		fs = root
	}
	index := 0
	files, _ := os.ReadDir(dirName)
	for index < len(files) {
		fileName = files[index]
		fi, _ = os.Stat(dirName + "/" + fileName.Name())
		dat, _ := os.ReadFile(dirName + "/" + fileName.Name())
		mode := fi.Mode()
		if mode.IsDir() {
			if fileName.Name() != "vendor" && fileName.Name() != ".git" {
				dirname := fileName.Name()
				fs.directories[dirname] = makeFilesystem(dirname, strings.ReplaceAll(dirName, "//", "/")+"/"+fileName.Name(), fs, fs.Filesystem)
				fs.MFS.Mkdir(filepath.Join(fs.rootPath, dirname), mode.Perm())

				replicateFilesystem(dirName+"/"+fileName.Name(), fs.directories[fileName.Name()])
			}

		} else {
			if fileName.Name() != "vendor" && fileName.Name() != ".git" {
				fs.files[fileName.Name()] = &file{
					name:     fileName.Name(),
					rootPath: strings.ReplaceAll(dirName, "//", "/") + "/" + fileName.Name(),
				}
				fname := fs.files[fileName.Name()].rootPath
				memfile, _ := fs.MFS.Create(fname)
				memfile.Truncate(fi.Size())
				memfile.Write(dat)
				fs.MFS.Chmod(filepath.Clean(fname), mode.Perm())
			}
		}

		index++
	}
	return fs
}

func makeFilesystem(dirName string, rootPath string, prev *filesystem, fsys *boot.Filesystem) *filesystem {
	fs := boot.InitFilesystem()
	if fsys == nil {
		fs = boot.InitFilesystem()
	} else {
		fs = fsys
	}

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

// Helper function to check file or dir exists
func (fs *filesystem) verifyPath(dirName string) (*filesystem, error) {
	checker := fs.handleRootNav(dirName)
	segments := strings.Split(dirName, "/")

	for _, segment := range segments {
		if segment == "." {
			continue
		}
		if len(segment) == 0 {
			continue
		}
		if segment == ".." {
			if checker.prev == nil {
				continue
			}
			checker = checker.prev
		} else if fs.doesDirExistRelativePath(segment, checker) {
			checker = checker.directories[segment]
		} else if fs.doesFileExistRelativePath(segment, checker) {
			return checker, nil
		} else {
			fmt.Printf("Error : %s doesn't exist\n", dirName)
			fmt.Println(segment)
			return fs, fmt.Errorf("Error : %s doesn't exist\n", dirName)
		}
	}
	return checker, nil
}

// Helper function to check file or dir exists
func (fs *filesystem) searchFS(dirName string) (*filesystem, error) {
	checker := root
	segments := strings.Split(dirName, "/")

	for _, segment := range segments {
		if segment == "." {
			continue
		}
		if len(segment) == 0 {
			continue
		}
		if segment == ".." {
			if checker.prev == nil {
				continue
			}
			checker = checker.prev
		} else if fs.doesDirExistRelativePath(segment, checker) {
			checker = checker.directories[segment]
		} else if fs.doesFileExistRelativePath(segment, checker) {
			return checker, nil
		}
	}
	return checker, nil
}

func (fs *filesystem) isDir(pathname string) bool {
	//if fs.rootPath == "." || path[0] != '/' {
	//	prefixPath = fs.rootPath + "/"
	//}
	return true
}

func (fs *filesystem) absPath(pathname string) string {
	var prefixPath string

	if pathname[0] != '/' {
		prefixPath = fs.rootPath + "/"
	} else {
		prefixPath = "."
	}

	absPath := prefixPath + pathname
	absPath = filepath.Clean(absPath)
	return absPath
}

func (fs *filesystem) handleRootNav(dirName string) *filesystem {
	if dirName[0] == '/' {
		return root
	}
	return fs
}

func (fs *filesystem) doesDirExistRelativePath(pathName string, fsys *filesystem) bool {
	if _, found := fsys.directories[pathName]; found {
		return true
	}
	return false
}

func (fs *filesystem) doesFileExistRelativePath(pathName string, fsys *filesystem) bool {
	if _, found := fsys.files[pathName]; found {
		return true
	}
	return false
}

func (fs *filesystem) doesDirExistAbsPath(pathName string) bool {
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

func (fs *filesystem) doesFileExistsAbsPath(pathName string) bool {
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
