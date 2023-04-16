package main

import (
	"os"
	"os/exec"
	"time"
)

func main() {
	cmd := exec.Command("/usr/local/bin/bot")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	println("start")
	err := cmd.Run()
	time.Sleep(time.Second * 2)
	println("end")
	if err != nil {
		println("error")
		panic(err)
	}
}
