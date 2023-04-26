package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/chzyer/readline"

	"github.com/marcellof23/vfs-TA/boot"
	"github.com/marcellof23/vfs-TA/constant"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResp struct {
	Data struct {
		Username string `json:"username"`
	} `json:"data"`
	Token string `json:"token"`
}

func authLoop() string {
	var input string
	line, err := readline.New(">")
	if err != nil {
		log.Fatal(err)
	}

	for {
		fmt.Println("Do you want to login or register (Type login or register to continue):")
		comm, err := line.Readline()
		if err != nil {
			log.Fatal(err)
		}

		if comm != "login" && comm != "register" {
			continue
		}

		input = comm
		break
	}
	return input
}

func register(dep *boot.Dependencies) (string, string) {
	var username, token string
	line, err := readline.New(">")
	if err != nil {
		log.Fatal(err)
	}

	var resp *http.Response
	for {
		fmt.Println("Please enter a Username:")
		uname, err := line.Readline()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Please enter a Password:")
		pass, err := line.Readline()
		if err != nil {
			log.Fatal(err)
		}
		values := Credentials{Username: uname, Password: pass}
		jsonValue, _ := json.Marshal(values)

		registerURL := constant.Protocol + dep.Config().Server.Addr + constant.ApiVer + "/user/sign-up"
		resp, err = http.Post(
			registerURL,
			"application/json",
			bytes.NewBuffer(jsonValue))
		if resp.StatusCode != http.StatusOK {
			fmt.Println("Enter a valid username and password	")
			continue
		}

		body, err := io.ReadAll(resp.Body)
		post := UserResp{}

		err = json.Unmarshal(body, &post)
		if err != nil {
			log.Printf("Reading body failed: %s", err)
			continue
		}

		username = post.Data.Username
		token = post.Token
		break
	}
	defer resp.Body.Close()
	return username, token
}

// login gets a custom Username from the current User.
func login(dep *boot.Dependencies) (string, string) {
	var username, token string
	line, err := readline.New(">")
	if err != nil {
		log.Fatal(err)
	}

	var resp *http.Response
	for {
		fmt.Println("Please enter a Username:")
		uname, err := line.Readline()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Please enter a Password:")
		pass, err := line.Readline()
		if err != nil {
			log.Fatal(err)
		}
		values := Credentials{Username: uname, Password: pass}
		jsonValue, _ := json.Marshal(values)

		loginURL := constant.Protocol + dep.Config().Server.Addr + constant.ApiVer + "/user/login"
		resp, err = http.Post(
			loginURL,
			"application/json",
			bytes.NewBuffer(jsonValue))
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Println("Username or password is invalid")
			continue
		}

		body, err := io.ReadAll(resp.Body)
		post := UserResp{}

		err = json.Unmarshal(body, &post)
		if err != nil {
			log.Printf("Reading body failed: %s", err)
			continue
		}

		username = post.Data.Username
		token = post.Token
		break
	}
	defer resp.Body.Close()
	return username, token
}

// initUser initializes the User object on startup.
func InitUser(dep *boot.Dependencies) *User {
	var username, token string
	command := authLoop()
	if command == "register" {
		username, token = register(dep)
	} else if command == "login" {
		username, token = login(dep)
	}

	currentUser := createUser(username, token)
	return currentUser
}
