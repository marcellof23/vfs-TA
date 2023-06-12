package fsys

import (
	"context"
	"fmt"
	"os"

	"github.com/marcellof23/vfs-TA/constant"
	"github.com/marcellof23/vfs-TA/pkg/model"
)

// Filesystem Commands

// Usage prints verifies that each command has the correct amount of
// command arguments and throws an error if not.
func (fs *Filesystem) Usage(comms []string) bool {
	switch comms[0] {
	case "mkdir":
		if len(comms) < 2 {
			fmt.Println(constant.UsageCommandMkdir)
			return false
		}
	case "pwd":
		if len(comms) > 1 {
			fmt.Println(constant.UsageCommandPwd)
			return false
		}
	case "ls":
		if len(comms) > 1 {
			fmt.Println(constant.UsageCommandLs)
			return false
		}
	case "cat":
		if len(comms) < 2 {
			fmt.Println(constant.UsageCommandCat)
			return false
		}
	case "stat":
		if len(comms) < 2 {
			fmt.Println(constant.UsageCommandStat)
			return false
		}
	case "rm":
		if len(comms) < 2 {
			fmt.Println(constant.UsageCommandRm)
			return false
		}
		if comms[1] == "-r" {
			if len(comms) < 3 {
				fmt.Println(constant.UsageCommandRm)
				return false
			}
		}
	case "cp":
		if len(comms) < 3 {
			fmt.Println(constant.UsageCommandCp)
			return false
		}
		if comms[1] == "-r" {
			if len(comms) < 4 {
				fmt.Println(constant.UsageCommandCp)
				return false
			}
		}
	case "chmod":
		if len(comms) < 3 {
			fmt.Println(constant.UsageCommandChmod)
			return false
		}
	case "upload":
		if len(comms) < 3 {
			fmt.Println(constant.UsageCommandUpload)
			return false
		}
		if comms[1] == "-r" {
			if len(comms) < 4 {
				fmt.Println(constant.UsageCommandUpload)
				return false
			}
		}
	case "migrate":
		if len(comms) < 3 {
			fmt.Println(constant.UsageCommandMigrate)
			return false
		}
	case "download":
		if len(comms) < 3 {
			fmt.Println(constant.UsageCommandDownload)
			return false
		}
	}
	return true
}

// Execute runs the commands passed into it.
func (fs *Filesystem) Execute(ctx context.Context, comms []string, publishing model.Publishing) (bool, error) {
	var err error
	if fs.Usage(comms) == false {
		return false, nil
	}

	role, ok := ctx.Value("role").(string)
	if !ok {
		return false, fmt.Errorf("User is not authorized!")
	}

	switch comms[0] {
	case "mkdir":
		err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.MkDir, ctx, publishing, comms[1])
	case "pwd":
		fs.Pwd()
	case "ls":
		fs.ListDir()
	case "cat":
		err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.Cat, ctx, publishing, comms[1])
	case "stat":
		stat, errs := fs.Stat(comms[1])
		if err == nil {
			fs.PrintStat(stat, comms[1])
		}
		err = errs
	case "rm":
		if comms[1] == "-r" {
			err = fs.FilesystemAccessAuth(ctx, role, true, comms[0], fs.RemoveDir, ctx, publishing, comms[2])
		} else {
			err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.RemoveFile, ctx, publishing, comms[1])
		}
	case "cp":
		if comms[1] == "-r" {
			err = fs.FilesystemAccessAuth(ctx, role, true, comms[0], fs.CopyDir, ctx, publishing, comms[2], comms[3])
		} else {
			err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.CopyFile, ctx, publishing, comms[1], comms[2])
		}
	case "chmod":
		err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.Chmod, ctx, publishing, comms[1], comms[2])
	case "chown":
		err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.Chown, ctx, publishing, comms[1], comms[2])
	case "upload":
		if comms[1] == "-r" {
			err = fs.FilesystemAccessAuth(ctx, role, true, comms[0], fs.UploadDir, ctx, publishing, comms[2], comms[3])
		} else {
			err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.UploadFile, ctx, publishing, comms[1], comms[2])
		}
	case "migrate":
		err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.Migrate, ctx, publishing, comms[1], comms[2])
	case "download":

		if comms[1] == "-r" {
			err = fs.FilesystemAccessAuth(ctx, role, true, comms[0], fs.DownloadRecursive, ctx, publishing, comms[2], comms[3])
		} else {
			err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.DownloadFile, ctx, publishing, comms[1], comms[2])
		}
	case "test":
		fs.Testing(ctx, comms[1])
	case "exit":
		fs.TearDown()
		os.Exit(1)
	default:
		return false, fmt.Errorf("%s: Command not found", comms[0])
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// Shell Commands

func (s *Shell) Usage(comms []string) bool {
	switch comms[0] {
	case "cd":
		if len(comms) != 2 {
			fmt.Println("Usage : cd [target directory]")
			return false
		}
	}
	return true
}

func (s *Shell) Execute(ctx context.Context, comms []string) bool {
	role, ok := ctx.Value("role").(string)
	if !ok {
		fmt.Println("User is not authorized!")
		return false
	}

	if s.Usage(comms) == false {
		return true
	}
	switch comms[0] {
	case "cd":
		err := s.Fs.FilesystemAccessAuth(ctx, role, false, comms[0], s.ChDir, ctx, false, comms[1])
		if err != nil {
			fmt.Println(err.Error())
		}
	case "clear":
		s.ClearScreen()
	default:
		return false
	}
	return true
}
