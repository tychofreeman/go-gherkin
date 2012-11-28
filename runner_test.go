package gherkin

import (
    "testing"
    . "github.com/tychofreeman/go-matchers"
    "io"
)

type MockScenario struct {
    rpt Report
}
func (ms MockScenario) AddStep(s step) {
}
func (ms MockScenario) Last() *step {
    return nil
}
func (ms MockScenario) Execute([]stepdef, io.Writer) Report {
    return ms.rpt
}
func (ms MockScenario) IsBackground() bool {
    return false
}

func TestReportsNumberOfScenarios(t *testing.T) {
    scenarios := []Scenario{
        MockScenario{rpt:Report{0,0,0,1,0,0}},
    }

    r := createWriterlessRunner()
    rpt := r.executeScenarios(scenarios)

    AssertThat(t, rpt.scenarioCount, Equals(1))
}

func TestReportsNumberOfStepsInScenarios(t *testing.T) {
    scenarios := []Scenario{
        MockScenario{rpt:Report{0,2,2,2,2,2}},
    }

    r := createWriterlessRunner()
    rpt := r.executeScenarios(scenarios)

    AssertThat(t, rpt.pendingSteps, Equals(2))
    AssertThat(t, rpt.skippedSteps, Equals(2))
    AssertThat(t, rpt.passedSteps, Equals(2))
    AssertThat(t, rpt.failedSteps, Equals(2))
}
