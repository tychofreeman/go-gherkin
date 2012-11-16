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
    Scenarios []string
}

func CreateRunner() *Runner {
    return &Runner{make([]stepdef, 1), 0, false, make([]string, 0)}
}

func (r *Runner) Register(pattern string, f func()) {
    r.steps = append(r.steps, createstep(pattern, f))
}

func (r *Runner) ExecuteFirstMatchingStep(line string) {
    for _, step := range r.steps {
        if step.execute(line) {
            r.StepCount++
            return
        }
    }
}

func (r *Runner) Step(line string) {
    defer func() {
        if rec := recover(); rec != nil {
            r.scenarioIsPending = true
        }
    }()

    givenMatch, _ := re.Compile(`(Given|When|Then|And|But|\*) (.*?)\s*$`)
    scenarioMatch, _ := re.Compile(`Scenario:\s*(.*?)\s*$`)
    if s := givenMatch.FindStringSubmatch(line); !r.scenarioIsPending && s != nil && len(s) > 1 {
        r.ExecuteFirstMatchingStep(s[2])
    } else if s := scenarioMatch.FindStringSubmatch(line); s != nil {
        r.Scenarios = append(r.Scenarios, s[1])
        r.scenarioIsPending = false
    }
}

func (r *Runner) Execute(file string) {
    lines := strings.Split(file, "\n")
    for _, line := range lines {
        r.Step(line)
    }
}

func Pending() {
    panic("Pending")
}
