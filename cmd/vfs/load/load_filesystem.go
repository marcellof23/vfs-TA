package load

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/marcellof23/vfs-TA/boot"
	"github.com/marcellof23/vfs-TA/constant"
	"github.com/marcellof23/vfs-TA/pkg/fsys"
	"github.com/marcellof23/vfs-TA/pkg/model"
)

func Untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}

			err = os.Chown(path, header.Uid, header.Gid)
			if err != nil {
				return err
			}
			continue
		}

		err = func() error {
			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				return err
			}
			defer file.Close()

			err = os.Chmod(path, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			err = os.Chown(path, header.Uid, header.Gid)
			if err != nil {
				return err
			}

			_, err = io.Copy(file, tarReader)
			if err != nil {
				return err
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

func LoadFilesystem(ctx context.Context, dep *boot.Dependencies, token string) error {
	syscall.Umask(0)

	backupURL := constant.Protocol + dep.Config().Server.Addr + constant.ApiVer + "/backup"

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, backupURL, nil)
	req.Header.Set("token", token)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return err
	}

	defer resp.Body.Close()

	log, ok := ctx.Value("server-logger").(*log.Logger)
	if !ok {
		return fmt.Errorf("ERROR: logger not initiated")
	}

	file, err := os.Create("backup.tar")
	if err != nil {
		log.Print("ERROR: ", err)
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Print("ERROR: ", err)
		return err
	}

	err = Untar("backup.tar", ".")
	if err != nil {
		log.Print("ERROR: ", err)
		return err
	}

	defer os.Remove("backup.tar")

	return nil
}

func LoadFilesystem2(ctx context.Context, dep *boot.Dependencies, token string) error {
	syscall.Umask(0)
	dst := "output"

	backupURL := constant.Protocol + dep.Config().Server.Addr + constant.ApiVer + "/backup"

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, backupURL, nil)
	req.Header.Set("token", token)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return err
	}

	defer resp.Body.Close()

	log, ok := ctx.Value("server-logger").(*log.Logger)
	if !ok {
		return fmt.Errorf("ERROR: logger not initiated")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print("ERROR: ", err)
		return err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		log.Print("ERROR: ", err)
		return err
	}

	for _, f := range zipReader.File {
		filePath := fsys.JoinPath(dst, f.Name)
		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			//fmt.Println("invalid file path")
			return fmt.Errorf("ERROR: invalid file path")
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, f.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), f.Mode()); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		fi, _ := dstFile.Stat()
		fmt.Println(fi.Name(), fi.Sys().(*syscall.Stat_t).Uid, fi.Sys().(*syscall.Stat_t).Gid)

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}

		dstFile.Close()
		fileInArchive.Close()
	}
	time.Sleep(100 * time.Second)
	return nil
}

func ReloadFilesys(ctx context.Context) error {
	log, ok := ctx.Value("server-logger").(*log.Logger)
	if !ok {
		return fmt.Errorf("ERROR: logger not initiated")
	}

	userState, ok := ctx.Value("userState").(model.UserState)
	if !ok {
		return errors.New("failed to get userState from context")
	}

	dep, ok := ctx.Value("dependency").(*boot.Dependencies)
	if !ok {
		return errors.New("failed to get dependency from context")
	}

	err := LoadFilesystem(ctx, dep, userState.Token)
	if err != nil {
		log.Println("ERROR: ", err)
		return errors.New("Load new filesysytem")
	}

	return nil
}
