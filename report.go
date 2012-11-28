package gherkin

type Report struct {
    scenarioCount int
    pendingSteps int
    skippedSteps int
    passedSteps int
    failedSteps int
}
