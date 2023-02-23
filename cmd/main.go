package main

import (
	"strings"
)

func shellLoop(currentUser *user) {
	var shellFlag bool

	//fs := boot.InitFilesystem()
	shell := initShell()
	fs := initFilesystem()
	prompt := currentUser.initPrompt()
	for {
		input, _ := prompt.Readline()
		input = strings.TrimSpace(input)
		if len(input) == 0 {
			continue
		}

		comms := strings.Split(input, " ")

		fs, shellFlag = shell.execute(comms, fs)
		if shellFlag == true {
			continue
		}

		if shellFlag == true {
			continue
		}
	}
}

//func init() {
//	var apiCmd = &cobra.Command{
//		Use:   "shell",
//		Short: "Runs the main Shell Loop for the Filesystem",
//		Run: func(cmd *cobra.Command, args []string) {
//			configfile := files
//			if len(args) != 0 {
//				configfile = args[0]
//			}
//			fmt.Println(configfile)
//
//			currentUser := initUser()
//
//			shellLoop(currentUser)
//		},
//	}
//
//	rootCmd.AddCommand(apiCmd)
//}

func main() {
	currentUser := initUser()
	shellLoop(currentUser)
}
