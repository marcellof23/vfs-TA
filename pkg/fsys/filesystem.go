package fsys

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/afero"

	"github.com/marcellof23/vfs-TA/boot"
	"github.com/marcellof23/vfs-TA/constant"
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

// Pwd prints pwd() the current working directory.
func (fs *Filesystem) Pwd() {
	fmt.Println(fs.rootPath)
}

// Stat gracefully ends the current session.
func (fs *Filesystem) Stat(filename string) (os.FileInfo, error) {
	path := filepath.Join(fs.rootPath, filename)
	info, err := fs.MFS.Stat(path)
	if err != nil {
		return nil, errors.New("file or Directory not found")
	}

	return info, nil
}

// UploadFile uploads a file to the virtual Filesystem.
func (fs *Filesystem) UploadFile(filename string) bool {
	fs.MFS.Create(fs.rootPath + "/" + filename)
	if _, exists := fs.files[filename]; exists {
		fmt.Printf("touch : file already exists")
		return false
	}
	newFile := &file{
		name:     filename,
		rootPath: fs.rootPath + "/" + filename,
	}
	fs.files[filename] = newFile
	return true
}

// UploadDir uploads a file to the virtual Filesystem.
func (fs *Filesystem) UploadDir(dirname string) bool {
	fs.MFS.Create(fs.rootPath + "/" + dirname)
	if _, exists := fs.files[dirname]; exists {
		fmt.Printf("touch : file already exists")
		return false
	}
	newFile := &file{
		name:     dirname,
		rootPath: fs.rootPath + "/" + dirname,
	}
	fs.files[dirname] = newFile
	return true
}

func (fs *Filesystem) Touch(filename string) error {
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
				errs := fmt.Sprintf("touch : directory with name %s already exists", segment)
				return errors.New(errs)
			}
		} else if !dirExist && idx < len(segments)-1 {
			errs := fmt.Sprintf("touch : cannot touch %s No such file or directory", segment)
			return errors.New(errs)
		} else if fileExist {
			errs := fmt.Sprintf("touch :  file with name %s already exists", segment)
			return errors.New(errs)
		} else if !fileExist && idx == len(segments)-1 {
			currFs.MFS.Create(currFs.rootPath + "/" + segment)
			if _, exists := currFs.files[segment]; exists {
				errs := fmt.Sprintf("touch : file %s  already exists", segment)
				return errors.New(errs)
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
func (fs *Filesystem) MkDir(dirName string) bool {
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
				fmt.Printf("mkdir : diRKrectory %s already exists", segment)
				return false
			}
		} else if !dirExist {
			err := fs.MFS.MkdirAll(filepath.Join(currFs.rootPath, segment), 0o700)
			if err != nil {
				fmt.Println(err)
				return false
			}

			newDir := &fileDir{
				name:        segment,
				rootPath:    filepath.Join(currFs.rootPath, segment),
				files:       make(map[string]*file),
				directories: make(map[string]*Filesystem),
				prev:        currFs,
			}

			currFs.directories[segment] = &Filesystem{currFs.MemFilesystem, newDir}
			currFs = currFs.directories[segment]
		}
	}

	return true
}

// RemoveFile removes a File from the virtual Filesystem.
func (fs *Filesystem) RemoveFile(filename string) error {
	absFilename := fs.absPath(filename)
	info, err := fs.Stat(filename)
	if err != nil {
		return errors.New("file or Directory does not exist")
	}

	if info.IsDir() {
		errs := fmt.Sprintf("rm : cannot remove '%s': Is a directory", filename)
		return errors.New(errs)
	}

	err = fs.MFS.Remove(absFilename)
	if err != nil {
		return err
	}

	fsTarget, err := fs.searchFS(absFilename)
	if err != nil {
		errs := fmt.Sprintf("rm : cannot remove '%s': path not found", filename)
		return errors.New(errs)
	}
	delete(fsTarget.files, filename)

	return nil
}

// RemoveDir removes a directory from the virtual Filesystem.
func (fs *Filesystem) RemoveDir(path string) error {
	path = fs.absPath(path)
	err := fs.MFS.RemoveAll(path)
	if err != nil {
		return err
	}

	walkFn := func(rootpath, path string, fs *Filesystem, err error) error {
		delete(fs.directories, path)
		return nil
	}

	err = walkDir(fs.directories[path], path, walkFn)
	if err != nil {
		fmt.Print(err)
	}
	delete(fs.directories, path)

	return nil
}

// CopyFile copy a file from source to destination on the virtual Filesystem.
func (fs *Filesystem) CopyFile(pathSource, pathDest string) error {
	fsDest, _ := fs.searchFS(pathDest)

	fl, err := fs.Stat(pathSource)
	if err != nil {
		fmt.Println("cp: ", err)
		return errors.New("file or Directory does not exist")
	}
	fmt.Println(fl.Name(), fl.Size())

	pathSourceFileName := fs.absPath(pathSource)
	pathTargetFileName := fs.absPath(pathDest)

	fsDest.Touch(pathDest)
	sourceFile, _ := fs.MFS.Open(pathSourceFileName)
	destFile, _ := fs.MFS.Open(pathTargetFileName)
	fmt.Println(destFile.Name())
	destFile.Truncate(fl.Size())

	b := make([]byte, fl.Size())
	sourceFile.Read(b)
	destFile.Write(b)

	return nil
}

// CopyDir copy a file from source to destination on the virtual Filesystem.
func (fs *Filesystem) CopyDir(pathSource, pathDest string) error {
	fsSource, _ := fs.searchFS(pathSource)
	fsDest, _ := fs.searchFS(pathDest)

	pathSource = filepath.Base(pathSource)
	pathDest = filepath.Base(pathDest)
	_, err := fs.Stat(pathSource)
	if err != nil {
		fmt.Println("cp: ", err)
		return errors.New("file or Directory does not exist")
	}
	fsDest.MkDir(pathDest)

	walkFn := func(rootPath, path string, _ *Filesystem, err error) error {
		if isDir, _ := fs.isDir(path); isDir {
			newDir := filepath.Join(pathDest, path)
			fsDest.MkDir(newDir)
		} else {
			newFile := filepath.Join(pathDest, rootPath, path)
			fsDest.Touch(newFile)

		}
		return nil
	}

	err = walkDir(fsSource, pathSource, walkFn)
	if err != nil {
		fmt.Print(err)
	}

	return nil
}

// ListDir lists a directory's contents.
func (fs *Filesystem) ListDir() {
	if fs.files != nil {
		for _, file := range fs.files {
			fmt.Println(file.name)
		}
	}
	if len(fs.directories) > 0 {
		for dirName := range fs.directories {
			coloredDir := fmt.Sprintf("\x1b[%dm%s\x1b[0m", constant.ColorBlue, dirName)
			fmt.Println(coloredDir)
		}
	}
}

func (fs *Filesystem) Chmod(perm, name string) error {
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

	return nil
}

func (fs *Filesystem) Cat(path string) {
	data, _ := afero.ReadFile(fs.MFS, path)
	fmt.Println(string(data))
}

func (fs *Filesystem) Testing(path string) {

	root.Touch("./a/b")
	fs, err := fs.searchFS(path)
	fmt.Println(fs.rootPath, err)
	fs.MFS.List()
	//walkFn := func(path string, fs *Filesystem, err error) error {
	//	fmt.Printf("%s \n", path)
	//	return nil
	//}
	//
	//walkDir(fs, ".", walkFn)
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
