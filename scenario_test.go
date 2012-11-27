package gherkin

import (
    "testing"
    . "github.com/tychofreeman/go-matchers"
)

func TestReportsNumberOfPendingSteps(t *testing.T) {
    scen := &scenario{}
    scen.AddStep(step{isPending:true})
    rpt := scen.Execute([]stepdef{}, nil)

    AssertThat(t, rpt.pendingSteps, Equals(1))
}

func TestReportsNumberOfSkippedSteps(t *testing.T) {
    scen := &scenario{}
    scen.AddStep(step{isPending:true})
    scen.AddStep(step{isPending:true})
    rpt := scen.Execute([]stepdef{}, nil)

    AssertThat(t, rpt.skippedSteps, Equals(1))
}
