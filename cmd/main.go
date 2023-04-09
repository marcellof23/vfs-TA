package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	fsys "github.com/marcellof23/vfs-TA/pkg/fsys"
	"github.com/marcellof23/vfs-TA/pkg/user"
)

func shellLoop(currentUser *user.User) {
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

		comms := strings.Split(input, " ")
		shellFlag = shells.Execute(comms)
		currentUser.SetPrompt(prompt, shells.Fs)
		if shellFlag {
			continue
		}
		Fsys = shells.Fs

		shellFlag = Fsys.Execute(comms)
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

			currentUser := user.InitUser()

			shellLoop(currentUser)
		},
	}

	rootCmd.AddCommand(apiCmd)
}
