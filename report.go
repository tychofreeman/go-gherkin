package gherkin

type Report struct {
    pendingSteps int
    skippedSteps int
    passedSteps int
    failedSteps int
}
