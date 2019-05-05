package main

import (
	"github.com/causton81/goslim/lib"
	"os"
	"os/exec"
)

func main() {
	workDir := os.Args[1]
	pkgToRun := os.Args[2]

	cmd := exec.Command("go", "run", pkgToRun)
	cmd.Dir = workDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	lib.Must(cmd.Run())
}
