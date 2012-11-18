package gherkin

import (
    "testing"
    . "github.com/tychofreeman/go-matchers"
)

var featureText = `Feature: My Feature
    Scenario: Scenario 1
        Given the first setup
        When the first action
        Then the first result
    Scenario: Scenario 2
        Given the second setup
        When the second action
        Then the second result
        And the other second result
    This is ignored`

func assertMatchCalledOrNot(t *testing.T, step string, pattern string, isCalled bool) {
        wasCalled := false
        f := func() {
            wasCalled = true
        }

        var g Runner
        g.Register(pattern, f)

        g.step(step)
        AssertThat(t, wasCalled, Equals(isCalled))
    }

func matchingFunctionIsCalled(t *testing.T, step string, pattern string) {
    assertMatchCalledOrNot(t, step, pattern, true)
}

func matchingFunctionIsNotCalled(t *testing.T, step string, pattern string) {
    assertMatchCalledOrNot(t, step, pattern, false)
}

func TestExecutesMatchingMethod(t *testing.T) {
    matchingFunctionIsCalled(t, "Given this step is called", ".")
}

func TestAvoidsNonMatchingMethod(t *testing.T) {
    matchingFunctionIsNotCalled(t, "Given this step is not called", "^A")
}

func TestCallsOnlyFirstMatchingMethod(t *testing.T) {
    wasCalled := false
    first := func() { }
    second := func() {
        wasCalled = true
    }

    var g Runner
    g.Register(".", first)
    g.Register(".", second)
    g.step("Given only the first step is called")
    AssertThat(t, wasCalled, Equals(false))
}

func TestRemovesGivenFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, "Given this is a given", "^this is a given$")
}

func TestRemovesWhenFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, "When this is a when", "^this is a when$")
}

func TestRemovesThenFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, "Then this is a then", "^this is a then$")
}

func TestRemovesAndFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, "And this is an and", "^this is an and$")
}

func TestRemovesButFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, "But this is a but", "^this is a but$")
}

func TestRemovesStarFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, "* this is a star", "^this is a star$")
}

func TestRemovesLeadingSpacesFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, "    Then we remove leading spaces", "^we remove leading spaces$")
}

func TestRemovesTrailingSpacesFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, "Then we remove trailing spaces   ", "^we remove trailing spaces$")
}

func TestMultipleStepsAreCalled(t *testing.T) {
    var g Runner

    firstWasCalled := false
    g.Register("^the first setup$", func() {
        firstWasCalled = true
    })

    secondWasCalled := false
    g.Register("^the first action$", func() {
        secondWasCalled = true
    })

    g.Execute(featureText)
    AssertThat(t, firstWasCalled, Equals(true))
    AssertThat(t, secondWasCalled, Equals(true))
}

func TestTellsNumberOfStepsExecuted(t *testing.T) {
    var g Runner

    g.Register("^the first setup$", func() {})
    g.Register("^the first action$", func() {})
    g.Register("^the first result$", func() {})

    g.Execute(featureText)
    AssertThat(t, g.StepCount, Equals(3))
}

func TestPendingSkipsTests(t *testing.T) {
    var g Runner

    g.Register("^the first setup$", func() { Pending() })
    actionWasCalled := false
    g.Register("^the first action$", func() { actionWasCalled = true })

    g.Execute(featureText)
    AssertThat(t, actionWasCalled, Equals(false))
}

func TestPendingDoesntSkipSecondScenario(t *testing.T) {
    var g Runner

    g.Register("^the first setup$", func() { Pending() })
    g.Register("^the second setup$", func() { } )
    secondActionCalled := false
    g.Register("^the second action$", func() { secondActionCalled = true })

    g.Execute(featureText)
    AssertThat(t, secondActionCalled, Equals(true))
}

func TestBackgroundIsRunBeforeEachScenario(t *testing.T) {
    var g Runner
    wasCalled := false
    g.Register("^background$", func() { wasCalled = true })
    g.Execute(`Feature: 
        Background:
            Given background
        Scenario:
            Then this
    `)

    AssertThat(t, wasCalled, IsTrue)
}

func TestCallsSeUptBeforeScenario(t *testing.T) {
    var g Runner
    setUpWasCalled := false
    g.SetSetUpFn(func() { setUpWasCalled = true })

    setUpCalledBeforeStep := false
    g.Register(".", func() { setUpCalledBeforeStep = setUpWasCalled })
    g.Execute(`Feature:
        Scenario:
            Then this`)

    AssertThat(t, setUpCalledBeforeStep, IsTrue)
}

func TestCallsTearDownBeforeScenario(t *testing.T) {
    var g Runner
    tearDownWasCalled := false
    g.SetTearDownFn(func() { tearDownWasCalled = true })

    g.Execute(`Feature:
        Scenario:
            Then this`)
    
    AssertThat(t, tearDownWasCalled, IsTrue)
}

// Need to introduce Backgrounds and Scenario Outlines/Examples and table inputs
// Need to support a lifecycle (setup/teardown) for each scenario executed.
// Support PyStrings?
// Support tags?
// Support reporting.
