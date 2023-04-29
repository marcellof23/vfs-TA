package user

import (
	"fmt"
	"log"

	"github.com/chzyer/readline"

	"github.com/marcellof23/vfs-TA/constant"
	"github.com/marcellof23/vfs-TA/pkg/fsys"
)

// The main User object.
type User struct {
	Username   string         // The User's onscreen name.
	Token      string         // User token
	Role       string         // User role
	accessList map[string]int // A map containing the unique hashes and access rights for each file.
}

// createUser creates a User object.
func createUser(username, role, token string) *User {
	return &User{
		Username: username,
		Role:     role,
		Token:    token,
	}
}

// updateUsername updates the name of the current User.
func (currentUser *User) SetPrompt(prompt *readline.Instance, fs *fsys.Filesystem) {
	var rootPath string
	if fs.GetRootPath() == "." {
		rootPath = "/"
	} else if fs.GetRootPath()[0] == '.' && fs.GetRootPath()[1] == '/' {
		rootPath = "/" + fs.GetRootPath()[2:]
	} else {
		rootPath = "/" + fs.GetRootPath()
	}

	coloredUsername := fmt.Sprintf("\x1b[%dm%s\x1b[0m", constant.ColorHiGreen, currentUser.Username)
	coloredRootPath := fmt.Sprintf("\x1b[%dm%s\x1b[0m", constant.ColorHiBlue, rootPath)
	prompt.SetPrompt(coloredUsername + ":" + coloredRootPath + "$> ")
}

// initPrompt initializes the input buffer for the
// shell.
func (currentUser *User) InitPrompt() *readline.Instance {
	autoCompleter := readline.NewPrefixCompleter(
		readline.PcItem("mkdir"),
		readline.PcItem("pwd"),
		readline.PcItem("ls"),
		readline.PcItem("cat"),
		readline.PcItem("stat"),
		readline.PcItem("touch"),
		readline.PcItem("rm"),
		readline.PcItem("cp"),
		readline.PcItem("mv"),
		readline.PcItem("chmod"),
		readline.PcItem("upload"),
		readline.PcItem("exit"),
	)

	coloredUsername := fmt.Sprintf("\x1b[%dm%s\x1b[0m", constant.ColorHiGreen, currentUser.Username)
	coloredRootPath := fmt.Sprintf("\x1b[%dm%s\x1b[0m", constant.ColorHiBlue, "/")
	prompt, err := readline.NewEx(&readline.Config{
		Prompt:          coloredUsername + ":" + coloredRootPath + "$>",
		HistoryFile:     "/tmp/commands.tmp",
		AutoComplete:    autoCompleter,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		log.Fatal(err)
	}
	return prompt
}
