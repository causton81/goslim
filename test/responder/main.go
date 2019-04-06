package main

import "github.com/causton81/goslim/slim"

type DummyDecisionTable struct {
}

func (*DummyDecisionTable) X() {}

type TestSlim struct {
}

func (*TestSlim) ThrowNormal() string {
	return "first"
}
func (*TestSlim) ThrowStopping() string {
	return "second"
}

func main() {
	slim.RegisterFixtureWithName(DummyDecisionTable{}, "fitnesse.slim.test.DummyDecisionTable")
	slim.RegisterFixtureWithName(TestSlim{}, "fitnesse.slim.test.TestSlim")
	slim.ListenAndServe()
}
