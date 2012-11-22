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
        But not the other first result
    Scenario: Scenario 2
        Given the second setup
        When the second action
        Then the second result
        And the other second result
    Scenario: Scenario 3
        * the third setup
        When     the third action has leading spaces
        When the third action has trailing spaces               
    This is ignored`

func assertMatchCalledOrNot(t *testing.T, step string, pattern string, isCalled bool) {
    wasCalled := false
    f := func(w World) {
        wasCalled = true
    }

    g := CreateRunner()
    g.RegisterStepDef(pattern, f)

    g.Execute(step)
    AssertThat(t, wasCalled, Equals(isCalled))
}

func matchingFunctionIsCalled(t *testing.T, step string, pattern string) {
    assertMatchCalledOrNot(t, step, pattern, true)
}

func matchingFunctionIsNotCalled(t *testing.T, step string, pattern string) {
    assertMatchCalledOrNot(t, step, pattern, false)
}

func TestExecutesMatchingMethod(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, ".")
}

func TestAvoidsNonMatchingMethod(t *testing.T) {
    matchingFunctionIsNotCalled(t, featureText, "^A")
}

func TestCallsOnlyFirstMatchingMethod(t *testing.T) {
    wasCalled := false
    first := func(w World) { }
    second := func(w World) {
        wasCalled = true
    }

    g := CreateRunner()
    g.RegisterStepDef(".", first)
    g.RegisterStepDef(".", second)
    g.Execute("Given only the first step is called")
    AssertThat(t, wasCalled, Equals(false))
}

func TestRemovesGivenFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the first setup$")
}

func TestRemovesWhenFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the first action$")
}

func TestRemovesThenFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the first result$")
}

func TestRemovesAndFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the other second result$")
}

func TestRemovesButFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^not the other first result$")
}

func TestRemovesStarFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the third setup$")
}

func TestRemovesLeadingSpacesFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the third action has leading spaces$")
}

func TestRemovesTrailingSpacesFromMatchLine(t *testing.T) {
    matchingFunctionIsCalled(t, featureText, "^the third action has trailing spaces$")
}

func TestMultipleStepsAreCalled(t *testing.T) {
    g := CreateRunner()

    firstWasCalled := false
    g.RegisterStepDef("^the first setup$", func(w World) {
        firstWasCalled = true
    })

    secondWasCalled := false
    g.RegisterStepDef("^the first action$", func(w World) {
        secondWasCalled = true
    })

    g.Execute(featureText)
    AssertThat(t, firstWasCalled, Equals(true))
    AssertThat(t, secondWasCalled, Equals(true))
}

func TestTellsNumberOfStepsExecuted(t *testing.T) {
    g := CreateRunner()

    g.RegisterStepDef("^the first setup$", func(w World) {})
    g.RegisterStepDef("^the first action$", func(w World) {})
    g.RegisterStepDef("^the first result$", func(w World) {})

    g.Execute(featureText)
    AssertThat(t, g.StepCount, Equals(3))
}

func TestPendingSkipsTests(t *testing.T) {
    g := CreateRunner()

    g.RegisterStepDef("^the first setup$", func(w World) { Pending() })
    actionWasCalled := false
    g.RegisterStepDef("^the first action$", func(w World) { actionWasCalled = true })

    g.Execute(featureText)
    AssertThat(t, actionWasCalled, Equals(false))
}

func TestPendingDoesntSkipSecondScenario(t *testing.T) {
    g := CreateRunner()

    g.RegisterStepDef("^the first setup$", func(w World) { Pending() })
    g.RegisterStepDef("^the second setup$", func(w World) { } )
    secondActionCalled := false
    g.RegisterStepDef("^the second action$", func(w World) { secondActionCalled = true })

    g.Execute(featureText)
    AssertThat(t, secondActionCalled, Equals(true))
}

func TestBackgroundIsRunBeforeEachScenario(t *testing.T) {
    g := CreateRunner()
    wasCalled := false
    g.RegisterStepDef("^background$", func(w World) { wasCalled = true })
    g.Execute(`Feature: 
        Background:
            Given background
        Scenario:
            Then this
    `)

    AssertThat(t, wasCalled, IsTrue)
}

func TestCallsSeUptBeforeScenario(t *testing.T) {
    g := CreateRunner()
    setUpWasCalled := false
    g.SetSetUpFn(func() { setUpWasCalled = true })

    setUpCalledBeforeStep := false
    g.RegisterStepDef(".", func(w World) { setUpCalledBeforeStep = setUpWasCalled })
    g.Execute(`Feature:
        Scenario:
            Then this`)

    AssertThat(t, setUpCalledBeforeStep, IsTrue)
}

func TestCallsTearDownBeforeScenario(t *testing.T) {
    g := CreateRunner()
    tearDownWasCalled := false
    g.SetTearDownFn(func() { tearDownWasCalled = true })

    g.Execute(`Feature:
        Scenario:
            Then this`)
    
    AssertThat(t, tearDownWasCalled, IsTrue)
}

func TestPassesTableListToMultiLineStep(t *testing.T) {
    g := CreateRunner()
    var data []map[string]string
    g.RegisterStepDef(".", func(w World) { data = w.multiStep })
    g.Execute(`Feature:
        Scenario:
            Then you should see these people
                |name|email|
                |Bob |bob@bob.com|
    `)

    expectedData := []map[string]string{
        map[string]string{"name":"Bob", "email":"bob@bob.com"},
    }
    AssertThat(t, data, Equals(expectedData))
}

func TestErrorsIfTooFewFieldsInMultiLineStep(t *testing.T) {
    g := CreateRunner()
    wasGivenRun := false
    wasThenRun := false
    // Assertions before end of test...
    defer func() {
        recover()
        AssertThat(t, wasGivenRun, IsFalse)
        AssertThat(t, wasThenRun, IsFalse)
    }()

    g.RegisterStepDef("given", func(w World) { wasGivenRun = true })
    g.RegisterStepDef("then", func(w World) { wasThenRun = true })

    g.Execute(`Feature:
        Scenario:
            Given given
                |name|addr|
                |bob|
            Then then`)
}

func TestSupportsMultipleMultiLineStepsPerScenario(t *testing.T) {
    g := CreateRunner()
    var givenData []map[string]string
    var whenData []map[string]string
    g.RegisterStepDef("given", func(w World) { givenData = w.multiStep })
    g.RegisterStepDef("when", func(w World) { whenData = w.multiStep })

    g.Execute(`Feature:
        Scenario:
            Given given
                |name|email|
                |Bob|bob@bob.com|
                |Jim|jim@jim.com|
            When when
                |breed|height|
                |wolf|2|
                |shihtzu|.5|
    `)

    expectedGivenData := []map[string]string{
        map[string]string{ "name":"Bob", "email":"bob@bob.com"},
        map[string]string{ "name":"Jim", "email":"jim@jim.com"},
    }

    expectedWhenData := []map[string]string{
        map[string]string{"breed":"wolf", "height":"2"},
        map[string]string{"breed":"shihtzu", "height":".5"},
    }
    
    AssertThat(t, givenData, Equals(expectedGivenData))
    AssertThat(t, whenData, Equals(expectedWhenData))
}

func TestAllowsAccessToFirstRegexCapture(t *testing.T) {
    g := CreateRunner()
    captured := ""
    g.RegisterStepDef("(thing)", func(w World) { captured = w.GetRegexParam() })
    g.Execute(`Feature:
        Scenario:
            Given thing
    `)

    AssertThat(t, captured, Equals("thing"))
}

func TestFailsGracefullyWithOutOfBoundsRegexCaptures(t *testing.T) {
    g := CreateRunner()
    g.RegisterStepDef(".", func(w World) { w.GetRegexParam() })

    defer func() {
        r := recover()
        AssertThat(t, r, Equals("GetRegexParam() called too many times."))
    }()

    g.Execute(`Feature:
        Scenario:
            Given .
    `)

}

func DISABLED_TestOnlyExecutesStepsBelowScenarioLine(t *testing.T) {
    g := CreateRunner()
    wasRun := false
    g.RegisterStepDef(".", func(w World) { wasRun = true })
    g.Execute(`Feature:
        Given .`)

    AssertThat(t, wasRun, IsFalse)
}

func TestScenarioOutlineWithoutExampleDoesNotExecute(t *testing.T) {
    g := CreateRunner()
    wasRun := false
    g.RegisterStepDef(".", func(w World) { wasRun = true})
    g.Execute(`Feature:
        Scenario Outline:
            Given .
    `)

    AssertThat(t, wasRun, IsFalse)
}

func DISABLED_TestExecutesScenarioOncePerLineInExample(t *testing.T) {
    g := CreateRunner()
    timesRun := 0
    g.RegisterStepDef(".", func(w World) { timesRun++ })
    g.Execute(`Feature:
        Scenario Outline:
            Given .
        Examples:
            |scenario num|
            |first|
            |second|
    `)

    AssertThat(t, timesRun, Equals(2))
}

// Need to introduce Scenario Outlines/Examples 
// Support PyStrings?
// Support tags?
// Support reporting.
