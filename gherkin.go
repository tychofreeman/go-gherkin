// Support the Gherkin language, as found in Ruby's Cucumber and Python's Lettuce projects.
package gherkin

import "io"
import matchers "github.com/tychofreeman/go-matchers"

// Static Runner object to make creating tests easier
var DefaultRunner = CreateRunner()

// Use this function to let the user know that this
// test is not complete.
func Pending() {
    panic("Pending")
}

// Pass-through for Runner.SetSetUpFn()
func SetUp(setup func()) {
    DefaultRunner.SetSetUpFn(setup)
}

// Pass-through for Runner.SetTearDownFn()
func TearDown(teardown func()) {
    DefaultRunner.SetTearDownFn(teardown)
}

// Pass-through for Runner.RegisterStepDef()
func RegisterStepDef(pattern string, stepdef func(*World)) {
    DefaultRunner.RegisterStepDef(pattern, stepdef)
}

func Given(pattern string, stepdef func(*World)) {
    DefaultRunner.RegisterStepDef(pattern, stepdef)
}

func When(pattern string, stepdef func(*World)) {
    DefaultRunner.RegisterStepDef(pattern, stepdef)
}

func Then(pattern string, stepdef func(*World)) {
    DefaultRunner.RegisterStepDef(pattern, stepdef)
}

func And(pattern string, stepdef func(*World)) {
    DefaultRunner.RegisterStepDef(pattern, stepdef)
}

// Pass-through for Runner.SetOutput()
func SetOutput(output io.Writer) {
    DefaultRunner.SetOutput(output)
}

// Pass-through for Runner.Run()
// This should be called after everything else.
func Run(t matchers.Errorable) {
    DefaultRunner.Run(t)
}
