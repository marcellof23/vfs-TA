package fsys

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/afero"

	"github.com/marcellof23/vfs-TA/boot"
	"github.com/marcellof23/vfs-TA/constant"
	"github.com/marcellof23/vfs-TA/pkg/producer"
)

type file struct {
	name     string // The name of the file.
	rootPath string // The absolute path of the file.
}

type fileDir struct {
	name        string                 // The name of the current directory we're in.
	rootPath    string                 // The absolute path to this directory.
	files       map[string]*file       // The list of files in this directory.
	directories map[string]*Filesystem // The list of directories in this directory.
	prev        *Filesystem            // a reference pointer to this directory's parent directory.
}

type Filesystem struct {
	*boot.MemFilesystem
	*fileDir
}

// Root node.
var root *Filesystem

// Getter and Setter Functions

// SetFilesystem set newFS to current Filesystem.
func (fs *Filesystem) SetFilesystem(newFS *Filesystem) {
	*fs = *newFS
}

func (fs *Filesystem) GetRootPath() string {
	return fs.rootPath
}

// Filesystem Library

// ChDir switches you to a different active directory.
func (fs *Filesystem) ChDir(_ context.Context, dirName string) {
	if dirName == "/" {
		fs = root
		return
	}

	fsVerified, err := fs.verifyPath(dirName)
	if err != nil {
		return
	}
	fs = fsVerified
}

// Pwd prints pwd() the current working directory.
func (fs *Filesystem) Pwd() {
	fmt.Println(fs.rootPath)
}

// Stat gracefully ends the current session.
func (fs *Filesystem) Stat(filename string) (os.FileInfo, error) {
	path := filepath.ToSlash(filepath.Join(fs.rootPath, filename))
	info, err := fs.MFS.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot stat %s: ", filename)
	}

	return info, nil
}

// UploadFile uploads a file to the virtual Filesystem.
func (fs *Filesystem) UploadFile(ctx context.Context, sourcePath, destPath string) error {
	fl, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("file %s not found", sourcePath)
	}

	dat, _ := os.ReadFile(sourcePath)
	mode := fl.Mode()

	fs.Touch(ctx, destPath)
	destFile, _ := fs.MFS.OpenFile(destPath, os.O_RDWR|os.O_CREATE, 0o600)
	destFile.Truncate(fl.Size())
	destFile.Write(dat)
	fs.MFS.Chmod(filepath.Clean(destPath), mode.Perm())

	token, err := GetTokenFromContext(ctx)
	if err != nil {
		return err
	}

	msg := producer.Message{
		Command:       "upload",
		Token:         token,
		AbsPathSource: destFile.Name(),
		AbsPathDest:   destFile.Name(),
		Buffer:        dat,
	}

	r := producer.Retry(producer.ProduceCommand, 3e9)
	go r(ctx, msg)

	return nil
}

// UploadDir uploads a file to the virtual Filesystem.
func (fs *Filesystem) UploadDir(ctx context.Context, sourcePath, destPath string) error {
	fsDest, _ := fs.searchFS(destPath)

	destPathBase := filepath.Base(destPath)
	fsDest.MkDir(ctx, destPathBase)
	fsDest = fsDest.directories[destPathBase]

	copyFilesystem(ctx, ".", sourcePath, fs.absPath(destPath), fsDest)
	return nil
}

func (fs *Filesystem) Touch(ctx context.Context, filename string) error {
	filename = fs.absPath(filename)
	currFs := root
	segments := strings.Split(filename, "/")
	for idx, segment := range segments {
		dirExist := fs.doesDirExistRelativePath(segment, currFs)
		fileExist := fs.doesFileExistRelativePath(segment, currFs)
		if segment == "." {
			continue
		}
		if len(segment) == 0 {
			continue
		}
		if segment == ".." {
			if currFs.prev == nil {
				continue
			}
			currFs = currFs.prev
		} else if dirExist {
			currFs = currFs.directories[segment]
			if idx == len(segments)-1 {
				return fmt.Errorf("touch : directory with name %s already exists", segment)
			}
		} else if !dirExist && idx < len(segments)-1 {
			return fmt.Errorf("touch : cannot touch %s No such file or directory", segment)
		} else if fileExist {
			return fmt.Errorf("touch :  file with name %s already exists", segment)
		} else if !fileExist && idx == len(segments)-1 {
			currFs.MFS.Create(currFs.rootPath + "/" + segment)
			if _, exists := currFs.files[segment]; exists {
				return fmt.Errorf("touch : file %s  already exists", segment)
			}
			newFile := &file{
				name:     segment,
				rootPath: currFs.rootPath + "/" + segment,
			}
			currFs.files[segment] = newFile

			return nil
		}
	}
	return nil
}

// MkDir makes a virtual directory.
func (fs *Filesystem) MkDir(ctx context.Context, dirName string) error {
	dirName = fs.absPath(dirName)
	currFs := root
	segments := strings.Split(dirName, "/")

	for idx, segment := range segments {
		dirExist := fs.doesDirExistRelativePath(segment, currFs)
		if segment == "." {
			continue
		}
		if len(segment) == 0 {
			continue
		}
		if segment == ".." {
			if currFs.prev == nil {
				continue
			}
			currFs = currFs.prev
		} else if dirExist {
			currFs = currFs.directories[segment]
			if idx == len(segments)-1 {
				return fmt.Errorf("mkdir : directory %s already exists", segment)
			}
		} else if !dirExist {
			err := fs.MFS.MkdirAll(filepath.ToSlash(filepath.Join(currFs.rootPath, segment)), 0o700)
			if err != nil {
				return err
			}

			newDir := &fileDir{
				name:        segment,
				rootPath:    filepath.ToSlash(filepath.Join(currFs.rootPath, segment)),
				files:       make(map[string]*file),
				directories: make(map[string]*Filesystem),
				prev:        currFs,
			}

			currFs.directories[segment] = &Filesystem{currFs.MemFilesystem, newDir}
			currFs = currFs.directories[segment]

			token, err := GetTokenFromContext(ctx)
			if err != nil {
				return err
			}

			msg := producer.Message{
				Command:       "mkdir",
				Token:         token,
				AbsPathSource: dirName,
				Buffer:        []byte{},
			}

			r := producer.Retry(producer.ProduceCommand, 3e9)
			go r(ctx, msg)
		}
	}

	return nil
}

// RemoveFile removes a File from the virtual Filesystem.
func (fs *Filesystem) RemoveFile(ctx context.Context, filename string) error {
	absFilename := fs.absPath(filename)

	info, err := fs.Stat(filename)
	_, err2 := fs.MFS.Stat(absFilename)
	if err != nil && err2 != nil {
		return errors.New("file or Directory does not exist")
	}

	if info.IsDir() {
		return fmt.Errorf("rm : cannot remove '%s': Is a directory", filename)
	}

	err = fs.MFS.Remove(absFilename)
	if err != nil {
		return err
	}

	baseFilename := filepath.Base(filename)
	fsTarget, err := fs.searchFS2(absFilename)
	if err != nil {
		errs := fmt.Sprintf("rm : cannot remove '%s': path not found", filename)
		return errors.New(errs)
	}
	delete(fsTarget.files, baseFilename)

	token, err := GetTokenFromContext(ctx)
	if err != nil {
		return err
	}

	msg := producer.Message{
		Command:       "rm",
		Token:         token,
		AbsPathSource: absFilename,
		AbsPathDest:   "",
		Buffer:        []byte{},
	}

	r := producer.Retry(producer.ProduceCommand, 3e9)
	go r(ctx, msg)

	return nil
}

// RemoveDir removes a directory from the virtual Filesystem.
func (fs *Filesystem) RemoveDir(ctx context.Context, dirname string) error {
	fsTarget, _ := fs.searchFS(dirname)
	dirname = fs.absPath(dirname)

	walkFn := func(rootpath, path string, fss *Filesystem, err error) error {
		fs.RemoveFile(ctx, filepath.ToSlash(filepath.Join(rootpath, path)))
		return nil
	}

	err := walkDir(fsTarget, dirname, walkFn)
	if err != nil {
		return err
	}
	baseDirName := filepath.Base(dirname)
	delete(fsTarget.prev.directories, baseDirName)

	err = fs.MFS.RemoveAll(dirname)
	if err != nil {
		return err
	}

	token, err := GetTokenFromContext(ctx)
	if err != nil {
		return err
	}

	msg := producer.Message{
		Command:       "rm -r",
		Token:         token,
		AbsPathSource: dirname,
		AbsPathDest:   "",
		Buffer:        []byte{},
	}

	r := producer.Retry(producer.ProduceCommand, 3e9)
	go r(ctx, msg)

	return nil
}

// CopyFile copy a file from source to destination on the virtual Filesystem.
func (fs *Filesystem) CopyFile(ctx context.Context, pathSource, pathDest string) error {
	flSource, err := fs.Stat(pathSource)
	if err != nil {
		return errors.New("cp: file or Directory source does not exist")
	}

	if flSource.IsDir() {
		return errors.New("cp: source path is a directory")
	}

	pathSourceFileName := fs.absPath(pathSource)
	pathTargetFileName := fs.absPath(pathDest)

	fs.Touch(ctx, pathDest)
	sourceFile, _ := fs.MFS.Open(pathSourceFileName)
	destFile, _ := fs.MFS.OpenFile(pathTargetFileName, os.O_RDWR|os.O_CREATE, 0o600)

	destFile.Truncate(flSource.Size())
	b := make([]byte, flSource.Size())
	sourceFile.Read(b)
	destFile.Write(b)

	token, err := GetTokenFromContext(ctx)
	if err != nil {
		return err
	}

	msg := producer.Message{
		Command:       "cp",
		Token:         token,
		AbsPathSource: pathSourceFileName,
		AbsPathDest:   pathTargetFileName,
		Buffer:        []byte{},
	}

	r := producer.Retry(producer.ProduceCommand, 3e9)
	go r(ctx, msg)

	return nil
}

// CopyDir copy a file from source to destination on the virtual Filesystem.
func (fs *Filesystem) CopyDir(ctx context.Context, pathSource, pathDest string) error {
	fsSource, _ := fs.searchFS(pathSource)
	fsDest, _ := fs.searchFS(pathDest)

	pathSource = filepath.Base(pathSource)
	pathDest = filepath.Base(pathDest)

	walkFn := func(rootPath, path string, _ *Filesystem, err error) error {
		if path == "" {
			return nil
		}
		if isDir, _ := fs.isDir(path); isDir {
			splitPaths := strings.Split(path, "/")
			splitPaths = splitPaths[1:]
			remainingPath := filepath.ToSlash(filepath.Join(splitPaths...))

			newDir := filepath.ToSlash(filepath.Join(pathDest, remainingPath))
			fsDest.MkDir(ctx, newDir)
		} else {
			splitPaths := strings.Split(rootPath, "/")
			splitPaths = splitPaths[1:]
			remainingPath := filepath.Join(splitPaths...)

			newFile := filepath.ToSlash(filepath.Join(pathDest, remainingPath, path))
			fs.CopyFile(ctx, filepath.ToSlash(filepath.Join(rootPath, path)), newFile)

		}
		return nil
	}

	err := walkDir(fsSource, pathSource, walkFn)
	if err != nil {
		return err
	}

	return nil
}

// Move copy a file from source to destination on the virtual Filesystem.
func (fs *Filesystem) Move(ctx context.Context, pathSource, pathDest string) error {
	pathSource = filepath.Base(pathSource)
	pathDest = filepath.Base(pathDest)

	fileSource, err := fs.Stat(pathSource)
	if err != nil {
		return errors.New("file or Directory does not exist")
	}

	infoDest, _ := fs.Stat(pathDest)
	if infoDest != nil {
		return errors.New("file or Directory destination already exist")
	}

	if !fileSource.IsDir() {
		err = fs.CopyFile(ctx, pathSource, pathDest)
		if err != nil {
			return errors.New("file or Directory does not exist")
		}
		fs.RemoveFile(ctx, pathSource)
	} else {
		err = fs.CopyDir(ctx, pathSource, pathDest)
		if err != nil {
			return errors.New("file or Directory does not exist")
		}
		fs.RemoveDir(ctx, pathSource)
	}

	return nil
}

// ListDir lists a directory's contents.
func (fs *Filesystem) ListDir() {
	if len(fs.files) > 0 {
		files := SortFiles(fs.files)
		for _, file := range files {
			fmt.Println(file)
		}
	}

	if len(fs.directories) > 0 {
		directories := SortDirs(fs.directories)
		for _, dirName := range directories {
			coloredDir := fmt.Sprintf("\x1b[%dm%s\x1b[0m", constant.ColorBlue, dirName)
			fmt.Println(coloredDir)
		}
	}
}

func (fs *Filesystem) Chmod(ctx context.Context, perm, name string) error {
	absName := fs.absPath(name)
	fs, err := fs.verifyPath(name)
	if err != nil {
		return errors.New("chmod: " + err.Error())
	}

	mode, err := strconv.ParseUint(perm, 8, 32)
	if err != nil {
		return errors.New("chmod: " + err.Error())
	}

	err = fs.MFS.Chmod(name, os.FileMode(mode))
	if err != nil {
		return errors.New("chmod: " + err.Error())
	}

	token, err := GetTokenFromContext(ctx)
	if err != nil {
		return err
	}

	msg := producer.Message{
		Command:       "chmod",
		Token:         token,
		AbsPathSource: absName,
		FileMode:      mode,
		AbsPathDest:   "",
		Buffer:        []byte{},
	}

	r := producer.Retry(producer.ProduceCommand, 3e9)
	go r(ctx, msg)

	return nil
}

func (fs *Filesystem) Cat(path string) error {
	path = fs.absPath(path)
	data, err := afero.ReadFile(fs.MFS, path)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

type MigrateResp struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func (fs *Filesystem) Migrate(ctx context.Context, pathSource, pathDest string) error {
	cli := []string{"gcs", "dos", "s3"}
	clientSource := pathSource
	if !contains(cli, clientSource) {
		return errors.New("source cloud storage is not supported or not found")
	}

	clientDest := pathDest
	if !contains(cli, clientDest) {
		return errors.New("destination cloud storage is not supported or not found")
	}

	if clientSource == clientDest {
		return errors.New("sis not supported or not foundource and destination cloud storage cannot be the same")
	}

	token, err := GetTokenFromContext(ctx)
	if err != nil {
		return err
	}

	// TODO: change localhost
	migrateURL := constant.Protocol + "localhost:8080" + constant.ApiVer + "/migrate"

	client := http.Client{}
	var param = url.Values{}
	param.Set("clientSource", clientSource)
	param.Set("clientDest", clientDest)

	var payload = bytes.NewBufferString(param.Encode())
	req, err := http.NewRequest(http.MethodPost, migrateURL, payload)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("token", token)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	post := MigrateResp{}

	err = json.Unmarshal(body, &post)
	if err != nil {
		return err
	}

	fmt.Println(post.Message)
	return nil
}

func (fs *Filesystem) Testing(path string) {

	if path == "hehe" {
		fs.MFS.List()
	}

	fmt.Println(fs.getAccess(path, "Normal"))
}

// SaveState aves the state of the VFS at this time.
func (fs *Filesystem) SaveState() {
	fmt.Println("Save the current state of the VFS")
}

// ReloadFilesys Resets the VFS and scraps all changes made up to this point.
// (basically like a rerun of InitFilesystem())
func (fs *Filesystem) ReloadFilesys() {
	fmt.Println("Refreshing...")
}

// TearDown gracefully ends the current session.
func (fs *Filesystem) TearDown() {
	fmt.Println("Teardown")
}
