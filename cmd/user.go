package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"log"

	"github.com/chzyer/readline"
)

// The main User object.
type User struct {
	userID     string         // A randomized hash string representing the users's unique ID.
	username   string         // The User's onscreen name.
	accessList map[string]int // A map containing the unique hashes and access rights for each file.
}

// generateRandomID generates a random userID value.
func generateRandomID() string {
	// return uint64(rand.Uint32()) << 32 + uint64(rand.Uint32())
	bytes := make([]byte, 64)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// createUser creates a User object.
func createUser(username string) *User {
	return &User{
		userID:   generateRandomID(),
		username: username,
	}
}

// updateUsername updates the name of the current User.
func (currentUser *User) updateUsername(username string) {
	currentUser.username = username
}

// initPrompt initializes the input buffer for the
// shell.
func (currentUser *User) initPrompt() *readline.Instance {
	autoCompleter := readline.NewPrefixCompleter(
		readline.PcItem("open"),
		readline.PcItem("close"),
		readline.PcItem("mkdir"),
		readline.PcItem("cd"),
		readline.PcItem("rmdir"),
		readline.PcItem("rm"),
		readline.PcItem("exit"),
	)
	prompt, err := readline.NewEx(&readline.Config{
		Prompt:          currentUser.username + "$>",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    autoCompleter,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		log.Fatal(err)
	}
	return prompt
}
