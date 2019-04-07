package main

import (
	"fmt"
	"github.com/causton81/goslim/slim"
)

type DummyDecisionTable struct {
}

func (*DummyDecisionTable) X() {}

type TestSlim struct {
	s string
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


func (*TestSlim) NullString() *string {
	return nil
}

func (*TestSlim) ReturnString() string {
	return "string"
}

func (ts *TestSlim) SetString(in string) {
	ts.s = in
}

func (ts *TestSlim) GetStringArg() string {
	return ts.s
}

type ExecuteThrowsReportableException struct {

}

func (*ExecuteThrowsReportableException) Execute() {
	panic(fmt.Errorf("A Reportable Exception"))
}

func main() {
	slim.SetTypePrefix("fitnesse.slim.test.TestSlim$")
	slim.RegisterFixtureWithName(DummyDecisionTable{}, "fitnesse.slim.test.DummyDecisionTable")
	slim.RegisterFixtureWithName(TestSlim{}, "fitnesse.slim.test.TestSlim")
	slim.RegisterFixtureWithName(ExecuteThrowsReportableException{}, "fitnesse.slim.test.ExecuteThrowsReportableException")
	slim.ListenAndServe()
}
