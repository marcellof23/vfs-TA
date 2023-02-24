package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func shellLoop(currentUser *User) {
	fmt.Println("asdfsdfs")
	shells := initShell()
	fs := initFilesystem()
	fmt.Println("asdfsdfs")
	prompt := currentUser.initPrompt()
	for {
		input, _ := prompt.Readline()
		input = strings.TrimSpace(input)
		if len(input) == 0 {
			continue
		}

		comms := strings.Split(input, " ")
		if comms[0] == "cd" {
			if len(comms) != 2 {
				fmt.Println("Usage : cd [directory]")
				continue
			}
			fs = shells.chDir(comms[1], fs)
		} else if comms[0] == "clear" {
			shells.clearScreen()
		} else {
			fs.execute(comms)
		}
	}
}

func init() {
	var apiCmd = &cobra.Command{
		Use:   "shell",
		Short: "Runs the main Shell Loop for the Filesystem",
		Run: func(cmd *cobra.Command, args []string) {
			configfile := files
			if len(args) != 0 {
				configfile = args[0]
			}
			fmt.Println(configfile)

			currentUser := initUser()

			shellLoop(currentUser)
		},
	}

	rootCmd.AddCommand(apiCmd)
}

//func main() {
//	currentUser := initUser()
//
//	shellLoop(currentUser)
//}
