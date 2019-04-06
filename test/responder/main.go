package main

import (
	"fmt"
	"github.com/causton81/goslim/slim"
)

type DummyDecisionTable struct {
}

func (*DummyDecisionTable) X() {}

type TestSlim struct {
}

type NoSuchConverter struct {
}

func (*TestSlim) SetNoSuchConverter(NoSuchConverter) {

}

func (*TestSlim) ThrowNormal() string {
	panic(fmt.Errorf("first"))
}
func (*TestSlim) ThrowStopping() string {
	panic(StopTest{msg: "second"})
}

type StopTest struct {
	msg string
}

func (st StopTest) Error() string {
	return st.msg
}

func main() {
	slim.SetTypePrefix("fitnesse.slim.test.TestSlim$")
	slim.RegisterFixtureWithName(DummyDecisionTable{}, "fitnesse.slim.test.DummyDecisionTable")
	slim.RegisterFixtureWithName(TestSlim{}, "fitnesse.slim.test.TestSlim")
	slim.ListenAndServe()
}
