package cmd

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/marcellof23/vfs-TA/boot"
	"github.com/marcellof23/vfs-TA/pkg/fsys"
	"github.com/marcellof23/vfs-TA/pkg/user"
)

func shellLoop(ctx context.Context, currentUser *user.User) {
	var shellFlag bool

	Fsys := fsys.New()
	prompt := currentUser.InitPrompt()
	shells := fsys.InitShell(Fsys)
	os.RemoveAll("output")

	for {
		input, _ := prompt.Readline()
		input = strings.TrimSpace(input)
		if len(input) == 0 {
			continue
		}

		commands := strings.Split(input, " ")

		// Execute the command for shell
		shellFlag = shells.Execute(ctx, commands)
		currentUser.SetPrompt(prompt, shells.Fs)

		Fsys = shells.Fs
		if shellFlag {
			continue
		}

		// Execute the command for filesystem
		shellFlag = Fsys.Execute(ctx, commands)
		currentUser.SetPrompt(prompt, Fsys)

		shells.SetFilesystem(Fsys)
		if shellFlag {
			continue
		}
	}
}

func init() {
	var apiCmd = &cobra.Command{
		Use:   "shell",
		Short: "Runs the main Shell Loop for the MemFilesystem",
		Run: func(cmd *cobra.Command, args []string) {
			LogFile, err := os.OpenFile("server-log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				log.Fatalf("error opening file: %v", err)
				return
			}
			defer LogFile.Close()

			configfile := files
			cfg, err := boot.LoadConfig(configfile)
			if err != nil {
				log.Fatal(err)
				return
			}

			dep, err := boot.InitDependencies(cfg)
			if err != nil {
				log.Fatal(err)
				return
			}

			logger := log.New(LogFile, time.Now().Format("2006-01-02 15:04:05")+": ", 0)
			ctx := context.WithValue(context.Background(), "server-logger", logger)

			currentUser := user.InitUser(dep)
			ctx = context.WithValue(ctx, "username", currentUser.Username)
			ctx = context.WithValue(ctx, "role", currentUser.Role)
			ctx = context.WithValue(ctx, "token", currentUser.Token)

			// err = LoadFilesystem(ctx, dep, currentUser.Token)
			// if err != nil {
			// 	logger.Println("ERROR: ", err)
			// 	return
			// }

			shellLoop(ctx, currentUser)
		},
	}

	rootCmd.AddCommand(apiCmd)
}
