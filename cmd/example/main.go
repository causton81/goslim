package main

import (
	"github.com/causton81/goslim/cmd/example/eg"
	"github.com/causton81/goslim/slim"
)

func main() {
	slim.RegisterFixture(eg.Division{})
	slim.ListenAndServe()
}


