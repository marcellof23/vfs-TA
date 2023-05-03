package main

import "time"

func a() string {
	time.Sleep(2 * time.Second)
	return "hallo"
}

func b() {
	go a()
}
func main() {
	//Fsys := fsys.New()

}
