package main

import (
	"github.com/causton81/goslim/cmd/example/eg"
	"github.com/causton81/goslim/slim"
	"github.com/causton81/goslim/test/responder"
)

func main() {
	slim.RegisterFixture(eg.Division{})
	slim.RegisterFixture(responder.TestSlim{})
	slim.ListenAndServe()
}


