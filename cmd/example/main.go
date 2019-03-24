package main

import (
	"github.com/causton/goslim/cmd/example/eg"
	"github.com/causton/goslim/slim"
)

func main() {
	slim.RegisterFixture(eg.Division{})
	slim.ListenAndServe()
}


