// Support the Gherkin language, as found in Ruby's Cucumber and Python's Lettuce projects.
package gherkin

import (
    re "regexp"
    "strings"
)

type stepdef struct {
    r *re.Regexp
    f func()
}

func createstep(p string, f func()) stepdef {
    r, _ := re.Compile(p)
    return stepdef{r, f}
}

func (s stepdef) execute(line string) bool {
    if s.r.MatchString(line) {
        s.f()
        return true
    }
    return false
}

type Runner struct {
    steps []stepdef
    StepCount int
    scenarioIsPending bool
    scenarios []string
    Features []string
    background []string
    collectBackground bool
}

func (r Runner) Scenarios() []string {
    return r.scenarios
}

// The recommended way to create a gherkin.Runner object.
func CreateRunner() *Runner {
    return &Runner{make([]stepdef, 1), 0, false, make([]string, 1), make([]string, 1), make([]string, 0), false}
}

// Register a step definition. This requires a regular expression
// pattern and a function to execute.
func (r *Runner) Register(pattern string, f func()) {
    r.steps = append(r.steps, createstep(pattern, f))
}

func (r *Runner) executeFirstMatchingStep(line string) {
    for _, step := range r.steps {
        if step.execute(line) {
            r.StepCount++
            return
        }
    }
}

func (r *Runner) step(line string) {
    defer func() {
        if rec := recover(); rec != nil {
            r.scenarioIsPending = true
        }
    }()

    givenMatch, _ := re.Compile(`(Given|When|Then|And|But|\*) (.*?)\s*$`)
    scenarioMatch, _ := re.Compile(`Scenario:\s*(.*?)\s*$`)
    featureMatch, _ := re.Compile(`Feature:\s*(.*?)\s*$`)
    backgroundMatch, _ := re.Compile(`Background:`)
    if s := givenMatch.FindStringSubmatch(line); !r.scenarioIsPending && s != nil && len(s) > 1 {
        if !r.collectBackground {
            r.executeFirstMatchingStep(s[2])
        } else {
            r.background = append(r.background, s[2])
        }
    } else if s := scenarioMatch.FindStringSubmatch(line); s != nil {
        r.collectBackground = false
        r.scenarios = append(r.scenarios, s[1])
        r.scenarioIsPending = false
        for _, bline := range r.background {
            r.executeFirstMatchingStep(bline)
        }
    } else if s := featureMatch.FindStringSubmatch(line); s != nil {
        r.Features = append(r.Features, s[1])
    } else if s := backgroundMatch.FindStringSubmatch(line); s != nil {
        r.collectBackground = true
    }
}

// Once the step definitions are Register()'d, use Execute() to
// parse and execute Gherkin data.
func (r *Runner) Execute(file string) {
    lines := strings.Split(file, "\n")
    for _, line := range lines {
        r.step(line)
    }
}

// Use this function to let the user know that this
// test is not complete.
func Pending() {
    panic("Pending")
}
