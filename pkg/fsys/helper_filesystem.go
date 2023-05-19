package fsys

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/marcellof23/vfs-TA/boot"
	"github.com/marcellof23/vfs-TA/constant"
	"github.com/marcellof23/vfs-TA/pkg/model"
	"github.com/marcellof23/vfs-TA/pkg/pubsub_notify/publisher"
)

type WalkDirFunc func(path, filename string, fs *Filesystem, err error) error

func GetUserStateFromContext(c context.Context) (model.UserState, error) {
	tmp := c.Value("userState")
	userState, ok := tmp.(model.UserState)
	if !ok {
		return model.UserState{}, constant.ErrUserStateNotFound
	}
	return userState, nil
}

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
		return []string{}, constant.ErrClientsNotFound
	}
	return clients, nil
}

func GetMaxFileSzFromContext(c context.Context) (int64, error) {
	tmp := c.Value("maxFileSize")
	maxSz, ok := tmp.(int64)
	if !ok {
		return 0, constant.ErrTokenNotFound
	}
	return maxSz, nil
}

func GetHostFromContext(c context.Context) (string, error) {
	tmp := c.Value("host")
	host, ok := tmp.(string)
	if !ok {
		return "", constant.ErrHostNotFound
	}
	return host, nil
}

func GetClientIDFromContext(c context.Context) (string, error) {
	tmp := c.Value("clientID")
	clientID, ok := tmp.(string)
	if !ok {
		return "", constant.ErrHostNotFound
	}
	return clientID, nil
}

func GetPublisherFromContext(c context.Context) (*publisher.Publisher, error) {
	tmp := c.Value("publisher")
	pubs, ok := tmp.(*publisher.Publisher)
	if !ok {
		return &publisher.Publisher{}, constant.ErrHostNotFound
	}
	return pubs, nil
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

		sz := info.Size()
		fname := filepath.Join(fs.rootPath, info.Name())
		if sz == 0 {
			sz = FileSizeMap[fname]
			fmt.Println("puntens")
		}

		fmt.Println("File: ", info.Name())
		fmt.Println("Size: ", sz)
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
			return fs, fmt.Errorf("Error : %s doesn't exist", dirName)
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
			return fs, fmt.Errorf("cannot stat '%s'", dirName)
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
			return fs, fmt.Errorf("cannot stat '%s'", dirName)
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

func GetFile(ctx context.Context, sourcePath string, targetFile *os.File) error {
	token, err := GetTokenFromContext(ctx)
	if err != nil {
		return err
	}

	dep, ok := ctx.Value("dependency").(*boot.Dependencies)
	if !ok {
		return errors.New("failed to get dependency from context")
	}

	filename := filepath.Clean(sourcePath)
	getFileURL := constant.Protocol + dep.Config().Server.Addr + constant.ApiVer + "/file/object?"

	client := http.Client{}
	var param = url.Values{}
	param.Add("filename", filename)

	req, err := http.NewRequest(http.MethodGet, getFileURL+param.Encode(), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", token)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return err
	}

	_, err = io.Copy(targetFile, resp.Body)
	if err != nil {
		fmt.Printf("Error downloading file: %s\n", err.Error())
	}

	return nil
}
