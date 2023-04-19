package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	fsys "github.com/marcellof23/vfs-TA/pkg/fsys"
	"github.com/marcellof23/vfs-TA/pkg/user"
)

func shellLoop(ctx context.Context, currentUser *user.User) {
	var shellFlag bool

	Fsys := fsys.New()
	prompt := currentUser.InitPrompt()
	shells := fsys.InitShell(Fsys)

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
			configfile := files
			if len(args) != 0 {
				configfile = args[0]
			}
			fmt.Println(configfile)

			errLogFile, err := os.OpenFile("error-log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				log.Fatalf("error opening file: %v", err)
			}
			defer errLogFile.Close()

			logger := log.New(errLogFile, "error: ", 0)
			ctx := context.WithValue(context.Background(), "logger", logger)

			currentUser := user.InitUser()

			shellLoop(ctx, currentUser)
		},
	}

	rootCmd.AddCommand(apiCmd)
}
