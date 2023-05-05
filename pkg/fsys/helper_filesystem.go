package fsys

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/marcellof23/vfs-TA/constant"
)

type WalkDirFunc func(path, filename string, fs *Filesystem, err error) error

func GetTokenFromContext(c context.Context) (string, error) {
	tmp := c.Value("token")
	token, ok := tmp.(string)
	if !ok {
		return "", constant.ErrTokenNotFound
	}
	return token, nil
}

func GetClientsFromContext(c context.Context) ([]string, error) {
	tmp := c.Value("clients")
	clients, ok := tmp.([]string)
	if !ok {
		return []string{}, constant.ErrTokenNotFound
	}
	return clients, nil
}

func GetHostFromContext(c context.Context) (string, error) {
	tmp := c.Value("host")
	host, ok := tmp.(string)
	if !ok {
		return "", constant.ErrTokenNotFound
	}
	return host, nil
}

func SortFiles(m map[string]*file) []string {
	keys := make([]string, 0, len(m))

	for key := range m {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return strings.ToLower(m[keys[i]].name) < strings.ToLower(m[keys[j]].name)
	})

	return keys
}

func SortDirs(m map[string]*Filesystem) []string {
	keys := make([]string, 0, len(m))

	for key := range m {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return strings.ToLower(m[keys[i]].name) < strings.ToLower(m[keys[j]].name)
	})

	return keys
}

func (fs *Filesystem) PrintStat(info *FileInfo, filename string) {
	if info != nil {
		var tipe string
		if info.IsDir() {
			tipe = "Directory"
		} else {
			tipe = "File"
		}

		fmt.Printf("%v", info.Sys())
		fmt.Println("File: ", info.Name())
		fmt.Println("Size: ", info.Size())
		fmt.Println("Access: ", info.Mode())
		fmt.Println("Modify: ", info.ModTime())
		fmt.Println("Type: ", tipe)
		fmt.Println("UserID: ", info.Uid)
		fmt.Println("GroupID: ", info.Gid)
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
func (fs *Filesystem) searchFS2(dirName string) (*Filesystem, error) {
	checker := root
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
	absPath = filepath.ToSlash(absPath)
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
	absPath = filepath.ToSlash(absPath)
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
		info, err := fs.MFS.Stat(filepath.ToSlash(filepath.Join(fs.rootPath, pathName)))
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
		info, err := fs.MFS.Stat(filepath.ToSlash(filepath.Join(fs.rootPath, pathName)))
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
			name1 := filepath.ToSlash(filepath.Join(path, dirName))
			if err := walkDir(fsys.directories[dirName], name1, walkDirFn); err != nil {
				return err
			}
		}
	}

	return nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func JoinPath(path ...string) string {
	joinedPath := filepath.Join(path...)
	unixPath := filepath.ToSlash(joinedPath)
	return unixPath
}
