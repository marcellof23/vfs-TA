package cmd

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/marcellof23/vfs-TA/boot"
	"github.com/marcellof23/vfs-TA/constant"
)

func LoadFilesystem(ctx context.Context, dep *boot.Dependencies, token string) error {
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
		filePath := filepath.Join(dst, f.Name)

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
	return nil
}
