package responder

import "fmt"

type DummyDecisionTable struct {
}

func (*DummyDecisionTable) X() {}

type TestSlim struct {
	s    string
	list []int
}

func (*TestSlim) Init() {

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

func (ts *TestSlim) OneList(in []int) {
	ts.list = in
}

func (ts *TestSlim) GetListArg() []int {
	return ts.list
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

type DecisionTableExecuteThrows struct {
}

func (*DecisionTableExecuteThrows) Execute() {
	panic("EXECUTE_THROWS")
}

func (*DecisionTableExecuteThrows) X() int {
	return 1
}

type DummyQueryTableWithNoTableMethod struct {

}

func (*DummyQueryTableWithNoTableMethod) Query() []int {
	return []int{}
}
