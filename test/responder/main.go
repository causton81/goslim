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
	panic(StopTest{detail: "second"})
}

type StopTest struct {
	detail  string
	message string
}

func (st StopTest) String() string {
	return st.message
}

func (st StopTest) Error() string {
	return st.detail
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

func (*TestSlim) ReturnInt() int {
	return 7
}

func (*TestSlim) EchoString(msg string) string {
	return msg
}

func (*TestSlim) ThrowStopTestExceptionWithMessage() {
	panic(StopTest{message: "Stop Test"})
}

func (*TestSlim) ThrowExceptionWithMessage() {
	panic("Test message")
}

type ExecuteThrowsReportableException struct {
}

func (*ExecuteThrowsReportableException) Execute() {
	panic(fmt.Errorf("A Reportable Exception"))
}

type TableFixture struct {
}

func (*TableFixture) Init(arg string) {

}

type DummyDecisionTableWithExecuteButNoReset struct {
}

func (*DummyDecisionTableWithExecuteButNoReset) Execute() {

}
func (*DummyDecisionTableWithExecuteButNoReset) X() int {
	return 1
}

func main() {
	slim.SetTypePrefix("fitnesse.slim.test.TestSlim$")
	slim.RegisterFixtureWithName(DummyDecisionTable{}, "fitnesse.slim.test.DummyDecisionTable")
	slim.RegisterFixtureWithName(TestSlim{}, "fitnesse.slim.test.TestSlim")
	slim.RegisterFixtureWithName(ExecuteThrowsReportableException{}, "fitnesse.slim.test.ExecuteThrowsReportableException")
	slim.RegisterFixtureWithName(TableFixture{}, "TableFixture")
	slim.RegisterFixtureWithName(DummyDecisionTableWithExecuteButNoReset{}, "fitnesse.slim.test.DummyDecisionTableWithExecuteButNoReset")
	slim.ListenAndServe()
}
