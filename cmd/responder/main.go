package main

import (
	"github.com/causton81/goslim/slim"
	"github.com/causton81/goslim/test/responder"
)

func main() {
	slim.SetTypePrefix("fitnesse.slim.test.TestSlim$")
	slim.RegisterFixtureWithName(responder.DummyDecisionTable{}, "fitnesse.slim.test.DummyDecisionTable")
	slim.RegisterFixtureWithName(responder.TestSlim{}, "fitnesse.slim.test.TestSlim")
	slim.RegisterFixtureWithName(responder.ExecuteThrowsReportableException{}, "fitnesse.slim.test.ExecuteThrowsReportableException")
	slim.RegisterFixtureWithName(responder.TableFixture{}, "TableFixture")
	slim.RegisterFixtureWithName(responder.DummyDecisionTableWithExecuteButNoReset{}, "fitnesse.slim.test.DummyDecisionTableWithExecuteButNoReset")
	slim.RegisterFixtureWithName(responder.DecisionTableExecuteThrows{}, "fitnesse.slim.test.DecisionTableExecuteThrows")
	slim.RegisterFixtureWithName(responder.DummyQueryTableWithNoTableMethod{}, "fitnesse.slim.test.DummyQueryTableWithNoTableMethod")

	slim.ListenAndServe()
}
