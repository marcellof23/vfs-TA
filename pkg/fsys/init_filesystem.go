package fsys

import (
	gofs "io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcellof23/vfs-TA/boot"
)

// Initiation Virtual MemFilesystem

func New() *Filesystem {
	// uncomment for recursively grab all files and directories from this level downwards.
	root = replicateFilesystem(".", ".", nil)

	// uncomment for initiate empty virtual Filesystem
	// root = makeFilesystem(".", ".", nil)

	fsys := root
	return fsys
}

// testFilessytemCreation initializes the Filesystem by replicating
// the current root directory and all it's child direcctories.
func copyFilesystem(dirName, replicatePath string, fs *Filesystem) *Filesystem {
	var fileName gofs.DirEntry
	var fi os.FileInfo

	index := 0

	files, _ := os.ReadDir(replicatePath)
	for index < len(files) {
		fileName = files[index]
		fi, _ = os.Stat(replicatePath + "/" + fileName.Name())
		dat, _ := os.ReadFile(replicatePath + "/" + fileName.Name())
		mode := fi.Mode()
		if mode.IsDir() {
			if fileName.Name() != "vendor" && fileName.Name() != ".git" {
				dirname := fileName.Name()
				fs.directories[dirname] = makeFilesystem(dirname, strings.ReplaceAll(dirName, "//", "/")+"/"+fileName.Name(), fs, fs.MemFilesystem)
				fs.MFS.Mkdir(filepath.Join(fs.rootPath, dirname), mode.Perm())
				replicateFilesystem(dirName+"/"+fileName.Name(), replicatePath+"/"+fileName.Name(), fs.directories[fileName.Name()])
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

// testFilessytemCreation initializes the Filesystem by replicating
// the current root directory and all it's child direcctories.
func replicateFilesystem(dirName, replicatePath string, fs *Filesystem) *Filesystem {
	var fileName gofs.DirEntry
	var fi os.FileInfo

	if dirName == "." {
		root = makeFilesystem(".", ".", nil, nil)
		fs = root
	}

	index := 0
	files, _ := os.ReadDir(replicatePath)
	for index < len(files) {
		fileName = files[index]
		fi, _ = os.Stat(replicatePath + "/" + fileName.Name())
		dat, _ := os.ReadFile(replicatePath + "/" + fileName.Name())
		mode := fi.Mode()
		if mode.IsDir() {
			if fileName.Name() != "vendor" && fileName.Name() != ".git" {
				dirname := fileName.Name()
				fs.directories[dirname] = makeFilesystem(dirname, strings.ReplaceAll(dirName, "//", "/")+"/"+fileName.Name(), fs, fs.MemFilesystem)
				fs.MFS.Mkdir(filepath.Join(fs.rootPath, dirname), mode.Perm())
				replicateFilesystem(dirName+"/"+fileName.Name(), replicatePath+"/"+fileName.Name(), fs.directories[fileName.Name()])
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

func makeFilesystem(dirName string, rootPath string, prev *Filesystem, fsys *boot.MemFilesystem) *Filesystem {
	fs := boot.InitFilesystem()
	if fsys == nil {
		fs = boot.InitFilesystem()
	} else {
		fs = fsys
	}

	return &Filesystem{
		fs,
		&fileDir{
			name:        dirName,
			rootPath:    rootPath,
			files:       make(map[string]*file),
			directories: make(map[string]*Filesystem),
			prev:        prev,
		},
	}
}