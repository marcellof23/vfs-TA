package fsys

import (
	"context"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/marcellof23/vfs-TA/constant"
)

func concludeAccess(accessSlice []string) string {
	r := true
	w := true
	x := true
	for _, v := range accessSlice {
		strAccess := strings.Split(v, "")
		r = r && (strAccess[0] == "r")
		w = w && (strAccess[1] == "w")
		x = x && (strAccess[2] == "x")
	}

	var access string
	if r {
		access += "r"
	} else {
		access += "-"
	}

	if w {
		access += "w"
	} else {
		access += "-"
	}

	if x {
		access += "x"
	} else {
		access += "-"
	}
	return access
}

func (fs *Filesystem) getAccess(path string, uid, gid int) (string, error) {
	checker := fs.handleRootNav(path)
	segments := strings.Split(path, "/")

	var accessSlice []string

	for _, segment := range segments {

		if segment == "." {
			rootStat, _ := fs.MFS.Stat(".")
			rootAccess := rootStat.Mode().Perm().String()

			if uid != fs.MFS.Uid(".") {
				rootAccess = rootAccess[len(rootAccess)-3:]
			} else {
				rootAccess = rootAccess[1:4]
			}
			accessSlice = append(accessSlice, rootAccess)

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
			dirName := filepath.ToSlash(filepath.Join(checker.rootPath, segment))
			dirStat, _ := fs.MFS.Stat(dirName)
			dirAccess := dirStat.Mode().Perm().String()

			if uid != fs.MFS.Uid(dirName) {
				dirAccess = dirAccess[len(dirAccess)-3:]
			} else {
				dirAccess = dirAccess[1:4]
			}
			accessSlice = append(accessSlice, dirAccess)
			checker = checker.directories[segment]
		} else if fs.doesFileExistRelativePath(segment, checker) {
			filename := filepath.ToSlash(filepath.Join(checker.rootPath, segment))
			fileStat, _ := fs.MFS.Stat(filename)
			fileAccess := fileStat.Mode().Perm().String()
			if uid != fs.MFS.Uid(filename) {
				fileAccess = fileAccess[len(fileAccess)-3:]
			} else {
				fileAccess = fileAccess[1:4]
			}
			accessSlice = append(accessSlice, fileAccess)
			acc := concludeAccess(accessSlice)
			if acc == "" {
				return "", constant.ErrUnauthorizedAccess
			}
			return acc, nil
		} else {
			acc := concludeAccess(accessSlice)
			if acc == "" {
				return "", constant.ErrUnauthorizedAccess
			}
			return acc, nil
		}
	}

	acc := concludeAccess(accessSlice)
	if acc == "" {
		return "", constant.ErrUnauthorizedAccess
	}
	return acc, nil
}

func checkAccess(access, requiredAccess string) bool {
	reqAccess := strings.Split(requiredAccess, "")
	acc := strings.Split(access, "")
	for i := 0; i < 3; i++ {
		if reqAccess[i] != "-" && acc[i] != reqAccess[i] {
			return false
		}
	}
	return true
}

func CallFunc(v reflect.Value, vargs []reflect.Value) error {
	res := v.Call(vargs)

	if res[0].Interface() == nil {
		return nil
	}

	err := res[0].Interface().(error)
	if err != nil {
		return err
	}
	return nil
}

func (fs *Filesystem) FilesystemAccessAuth(ctx context.Context, role string, isRec bool, command string, f interface{}, args ...interface{}) error {
	v := reflect.ValueOf(f)
	vargs := make([]reflect.Value, len(args))
	for i, arg := range args {
		vargs[i] = reflect.ValueOf(arg)
	}

	var comms = make([]string, 3)
	sourceOnly := false
	comms[1] = args[2].(string)
	if len(args) > 3 {
		comms[2] = args[3].(string)
	} else {
		sourceOnly = true
	}

	userState, err := GetUserStateFromContext(ctx)
	if err != nil {
		return err
	}

	if _, ok := constant.Command[command]; ok {
		var srcPath, dstPath string
		srcPath = comms[1]
		if !sourceOnly {
			dstPath = comms[2]
		}

		var srcAccess, dstAccess string
		srcAccess = "---"
		dstAccess = "---"

		srcAccess, err := fs.getAccess(srcPath, userState.UserID, userState.GroupID)
		if err != nil {
			return err
		}

		if !sourceOnly {
			dstAccess, err = fs.getAccess(dstPath, userState.UserID, userState.GroupID)
			if err != nil && dstPath != "" {
				return err
			}
		}

		switch command {
		case "cp":
			if isRec {
				err = fs.CheckCPRecPath(srcPath, dstPath)
				if err != nil {
					return err
				}

				if role == "Normal" {
					if !checkAccess(srcAccess, "r-x") {
						return constant.ErrUnauthorizedAccess
					}

					if !checkAccess(dstAccess, "rw-") {
						return constant.ErrUnauthorizedAccess
					}
				}

			} else {
				err = fs.CheckCPPath(srcPath, dstPath)
				if err != nil {
					return err
				}

				if role == "Normal" {
					if !checkAccess(srcAccess, "r--") {
						return constant.ErrUnauthorizedAccess
					}

					if !checkAccess(dstAccess, "-w-") {
						return constant.ErrUnauthorizedAccess
					}
				}

			}
		case "rm":
			if role == "Normal" {
				if !checkAccess(srcAccess, "-wx") {
					return constant.ErrUnauthorizedAccess
				}
			}
		case "upload":
			if isRec {
				err = fs.CheckUploadRecPath(srcPath, dstPath)
				if err != nil {
					return err
				}
				if role == "Normal" {
					if !checkAccess(dstAccess, "-wx") {
						return constant.ErrUnauthorizedAccess
					}
				}
			} else {
				err = fs.CheckUploadPath(srcPath, dstPath)
				if err != nil {
					return err
				}
				if role == "Normal" {
					if !checkAccess(dstAccess, "-w-") {
						return constant.ErrUnauthorizedAccess
					}
				}
			}
		case "mkdir":
			if role == "Normal" {
				splitPaths := strings.Split(srcPath, "/")
				splitPaths = splitPaths[:len(splitPaths)-1]
				remainingSourceDest := filepath.ToSlash(filepath.Join(splitPaths...))
				if len(splitPaths) > 1 {
					srcAccess, err = fs.getAccess(remainingSourceDest, userState.UserID, userState.GroupID)
					if err != nil {
						return err
					}
				}

				srcAccess, err = fs.getAccess(".", userState.UserID, userState.GroupID)
				if err != nil {
					return err
				}

				if !checkAccess(srcAccess, "-wx") {
					return constant.ErrUnauthorizedAccess
				}
			}
		case "cat":
			if role == "Normal" {
				if !checkAccess(srcAccess, "r--") {
					return constant.ErrUnauthorizedAccess
				}
			}
		case "cd":
			if role == "Normal" {
				if !checkAccess(srcAccess, "--x") {
					return constant.ErrUnauthorizedAccess
				}
			}
		case "download":
			if role == "Normal" {
				if !checkAccess(srcAccess, "r--") {
					return constant.ErrUnauthorizedAccess
				}
			}
		case "chmod":
			if role == "Normal" {
				if userState.UserID != fs.MFS.Uid(srcPath) {
					return constant.ErrUnauthorizedAccess
				}
			}
		case "migrate":
			if role == "Normal" {
				return constant.ErrUnauthorizedAccess
			}
		}
	}

	return CallFunc(v, vargs)
}
