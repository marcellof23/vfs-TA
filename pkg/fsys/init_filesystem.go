package fsys

import (
	"context"
	"fmt"
	gofs "io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/marcellof23/vfs-TA/boot"
	"github.com/marcellof23/vfs-TA/pkg/cache"
	"github.com/marcellof23/vfs-TA/pkg/model"
	"github.com/marcellof23/vfs-TA/pkg/producer"
)

// Initiation Virtual MemFilesystem

func New() *model.Filesystem {
	cache.LruCache = cache.Constructor()

	// uncomment for recursively grab all files and directories from this level downwards.
	root = ReplicateFilesystem(".", "backup", nil)

	// uncomment for initiate empty virtual Filesystem
	// root = makeFilesystem(".", ".", nil)

	fsys := root
	return fsys
}

// testFilessytemCreation initializes the Filesystem by replicating
// the current root directory and all it's child direcctories.
func copyFilesystem(ctx context.Context, dirName, replicatePath, targetPath string, fs *model.Filesystem, idx int) *model.Filesystem {
	var fileName gofs.DirEntry
	var fi os.FileInfo

	userState, ok := ctx.Value("userState").(model.UserState)
	if !ok {
		return &model.Filesystem{}
	}

	index := 0

	files, _ := os.ReadDir(replicatePath)
	for index < len(files) {
		fileName = files[index]
		fi, _ = os.Stat(replicatePath + "/" + fileName.Name())
		dat, _ := os.ReadFile(replicatePath + "/" + fileName.Name())
		mode := fi.Mode()
		if mode.IsDir() {
			dirname := fileName.Name()
			fs.MkDir(ctx, dirname)
			fs.MFS.Chown(filepath.ToSlash(filepath.Clean(dirname)), userState.UserID, userState.GroupID)
			copyFilesystem(ctx, dirName+"/"+fileName.Name(), replicatePath+"/"+fileName.Name(), targetPath, fs.Directories[fileName.Name()], idx+1)
		} else {
			fs.Files[fileName.Name()] = &model.File{
				Name:     fileName.Name(),
				RootPath: strings.ReplaceAll(dirName, "//", "/") + "/" + fileName.Name(),
			}
			fname := fs.Files[fileName.Name()].RootPath
			memfile, _ := fs.MFS.Create(filepath.ToSlash(filepath.Join(targetPath, fname)))
			memfile.Truncate(fi.Size())
			memfile.Write(dat)
			fs.MFS.Chmod(filepath.ToSlash(filepath.Clean(fname)), mode.Perm())
			fs.MFS.Chown(filepath.ToSlash(filepath.Clean(fname)), userState.UserID, userState.GroupID)

			token, err := GetTokenFromContext(ctx)
			if err != nil {
				fmt.Println(err)
			}

			msg := producer.Message{
				Command:       "upload",
				Token:         token,
				AbsPathSource: fname,
				AbsPathDest:   targetPath,
				Buffer:        dat,
				FileMode:      uint64(mode),
				Offset:        idx,
				Uid:           userState.UserID,
				Gid:           userState.GroupID,
			}

			r := producer.Retry(producer.ProduceCommand, 3e9)
			go r(ctx, msg)
		}

		index++
	}
	return fs
}

// testFilessytemCreation initializes the Filesystem by replicating
// the current root directory and all it's child direcctories.
func ReplicateFilesystem(dirName, replicatePath string, fs *model.Filesystem) *model.Filesystem {
	var fileName gofs.DirEntry
	var fi os.FileInfo

	if dirName == "." {
		root = makeFilesystem(".", ".", nil, nil)
		fs = root
	}

	fl, _ := os.Stat("backup")
	fs.MFS.Chmod("/", fl.Mode())
	index := 0
	files, _ := os.ReadDir(replicatePath)
	for index < len(files) {
		fileName = files[index]
		fi, _ = os.Stat(replicatePath + "/" + fileName.Name())
		dat, _ := os.ReadFile(replicatePath + "/" + fileName.Name())
		mode := fi.Mode()
		if mode.IsDir() {
			dirname := fileName.Name()
			fs.Directories[dirname] = makeFilesystem(dirname, strings.ReplaceAll(dirName, "//", "/")+"/"+fileName.Name(), fs, fs.MemFilesystem)

			dirNameClean := filepath.ToSlash(filepath.Join(fs.RootPath, dirname))
			fs.MFS.Mkdir(dirNameClean, mode)
			fs.MFS.Chown(dirNameClean, int(fi.Sys().(*syscall.Stat_t).Uid), int(fi.Sys().(*syscall.Stat_t).Gid))
			ReplicateFilesystem(dirName+"/"+fileName.Name(), replicatePath+"/"+fileName.Name(), fs.Directories[fileName.Name()])
		} else {
			fs.Files[fileName.Name()] = &model.File{
				Name:     fileName.Name(),
				RootPath: strings.ReplaceAll(dirName, "//", "/") + "/" + fileName.Name(),
			}
			fname := fs.Files[fileName.Name()].RootPath
			memfile, _ := fs.MFS.Create(fname)
			memfile.Truncate(fi.Size())
			memfile.Write(dat)

			fnameClean := filepath.ToSlash(filepath.Clean(fname))
			fs.MFS.Chmod(fnameClean, mode)
			fs.MFS.Chown(fnameClean, int(fi.Sys().(*syscall.Stat_t).Uid), int(fi.Sys().(*syscall.Stat_t).Gid))
			cache.LruCache.Put(fnameClean, fi.Size(), fs)

		}
		index++
	}
	return fs
}

func makeFilesystem(dirName string, rootPath string, prev *model.Filesystem, fsys *boot.MemFilesystem) *model.Filesystem {
	fs := boot.InitFilesystem()
	if fsys == nil {
		fs = boot.InitFilesystem()
	} else {
		fs = fsys
	}

	return &model.Filesystem{
		fs,
		&model.FileDir{
			Name:        dirName,
			RootPath:    rootPath,
			Files:       make(map[string]*model.File),
			Directories: make(map[string]*model.Filesystem),
			Prev:        prev,
		},
	}
}
