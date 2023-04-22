package fsys

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrTokenNotFound = errors.New("failed to get token from context")

type WalkDirFunc func(path, filename string, fs *Filesystem, err error) error

func GetTokenFromContext(c context.Context) (string, error) {
	tmp := c.Value("token")
	token, ok := tmp.(string)
	if !ok {
		return "", ErrTokenNotFound
	}
	return token, nil
}

func (fs *Filesystem) PrintStat(info os.FileInfo, filename string) {
	if info != nil {
		var tipe string
		if info.IsDir() {
			tipe = "Directory"
		} else {
			tipe = "File"
		}
		fmt.Println("File: ", info.Name())
		fmt.Println("Size: ", info.Size())
		fmt.Println("Access: ", info.Mode())
		fmt.Println("Modify: ", info.ModTime())
		fmt.Println("Type: ", tipe)
	}
}

// verifyPath is a function to check file or dir exists
func (fs *Filesystem) verifyPath(dirName string) (*Filesystem, error) {
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
			return fs, fmt.Errorf("Error : %s doesn't exist\n", dirName)
		}
	}
	return checker, nil
}

// searchFS to check file or dir exists
func (fs *Filesystem) searchFS(dirName string) (*Filesystem, error) {
	checker := fs.handleRootNav(dirName)
	segments := strings.Split(dirName, "/")

	for idx, segment := range segments {
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
		} else if idx != len(segments)-1 {
			return fs, fmt.Errorf("Error : %s doesn't exist\n", dirName)
		}
	}
	return checker, nil
}

func (fs *Filesystem) isDir(pathname string) (bool, error) {
	var prefixPath string

	if pathname[0] != '/' {
		prefixPath = fs.rootPath + "/"
	} else {
		prefixPath = "."
	}

	absPath := prefixPath + pathname
	absPath = filepath.Clean(absPath)
	info, err := fs.MFS.Stat(absPath)
	if err != nil {
		return false, err
	}

	if info.IsDir() {
		return true, nil
	} else {
		return false, nil
	}

}
func (fs *Filesystem) absPath(pathname string) string {
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

func (fs *Filesystem) handleRootNav(dirName string) *Filesystem {
	if dirName[0] == '/' {
		return root
	}
	return fs
}

func (fs *Filesystem) doesDirExistRelativePath(pathName string, fsys *Filesystem) bool {
	if _, found := fsys.directories[pathName]; found {
		return true
	}
	return false
}

func (fs *Filesystem) doesFileExistRelativePath(pathName string, fsys *Filesystem) bool {
	if _, found := fsys.files[pathName]; found {
		return true
	}
	return false
}

func (fs *Filesystem) doesDirExistAbsPath(pathName string) bool {
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

func (fs *Filesystem) doesFileExistsAbsPath(pathName string) bool {
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

func walkDir(fsys *Filesystem, path string, walkDirFn WalkDirFunc) error {
	err := walkDirFn(fsys.rootPath, path, fsys, nil)
	if err != nil {
		return err
	}

	if fsys.files != nil {
		for _, fl := range fsys.files {
			walkDirFn(path, fl.name, fsys, nil)
		}
	}

	if len(fsys.directories) > 0 {
		for dirName := range fsys.directories {
			name1 := filepath.Join(path, dirName)
			if err := walkDir(fsys.directories[dirName], name1, walkDirFn); err != nil {
				return err
			}
		}
	}

	return nil
}
