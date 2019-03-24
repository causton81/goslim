package main

import (
	"os"
	"os/exec"
	"syscall"
)
import "github.com/causton81/goslim/lib"

func main() {
	lib.Must(os.Chdir(os.Args[1]))
	cmd, err := exec.LookPath("go")
	lib.Must(err)
	lib.Must(syscall.Exec(cmd, []string{cmd, "run", os.Args[2]}, os.Environ()))
}
