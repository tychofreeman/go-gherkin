// Support the Gherkin language, as found in Ruby's Cucumber and Python's Lettuce projects.
package gherkin

import (
    re "regexp"
    "strings"
    "fmt"
)

// Passed to each step-definition
type World struct {
    regexParams []string
    regexParamIndex int
    multiStep []map[string]string
}

// Allows access to step definition regular expression captures.
func (w World) GetRegexParam() string {
    w.regexParamIndex++
    if w.regexParamIndex >= len(w.regexParams) {
        panic("GetRegexParam() called too many times.")
    }
    return w.regexParams[w.regexParamIndex]
}

type stepdef struct {
    r *re.Regexp
    f func(World)
}

func createstep(p string, f func(World)) stepdef {
    r, _ := re.Compile(p)
    return stepdef{r, f}
}

func (s stepdef) execute(line step, mlData []map[string]string) bool {
    if s.r.MatchString(line.String()) {
        if s.f != nil {
            substrs := s.r.FindStringSubmatch(line.String())
            s.f(World{regexParams:substrs, multiStep:line.mldata})
        }
        return true
    }
    return false
}

type step struct {
    line string
    keys []string
    mldata []map[string]string
}
func (s step) String() string {
    return s.line
}
func StepFromString(in string) step{
    return step{ line : in, keys: []string{}, mldata : []map[string]string{} }
}
func (s *step) addMlData(line map[string]string) {
    s.mldata = append(s.mldata, line)
}

func (s *step) setMlKeys(keys []string) {
    s.keys = keys
}

type scenario []step

type Runner struct {
    steps []stepdef
    StepCount int
    scenarioIsPending bool
    background scenario
    collectBackground bool
    setUp func()
    tearDown func()
    keys []string
    mlStep []map[string]string
    currScenario scenario
    lastExecutedIndex int
    scenarios []scenario
}

func (r *Runner) addStepLine(line string) {
    r.currScenario = append(r.currScenario, StepFromString(line))
}

func (r *Runner) currStepLine() step {
    if len(r.currScenario) > 0 {
        return r.currScenario[len(r.currScenario) - 1]
    }
    return StepFromString("")
}

func (r *Runner) resetStepLine() {
    r.lastExecutedIndex = len(r.currScenario)
}

func (r *Runner) hasOutstandingStep() bool {
    return r.lastExecutedIndex < len(r.currScenario)
}

// Register a set-up function to be called at the beginning of each scenario
func (r *Runner) SetSetUpFn(setUp func()) {
    r.setUp = setUp
}

// Register a tear-down function to be called at the end of each scenario
func (r *Runner) SetTearDownFn(tearDown func()) {
    r.tearDown = tearDown
}

// The recommended way to create a gherkin.Runner object.
func CreateRunner() *Runner {
    return &Runner{[]stepdef{}, 0, false, scenario{}, false, nil, nil, nil, []map[string]string{}, nil, -1, []scenario{}}
}

// Register a step definition. This requires a regular expression
// pattern and a function to execute.
func (r *Runner) RegisterStepDef(pattern string, f func(World)) {
    r.steps = append(r.steps, createstep(pattern, f))
}

func (r *Runner) reset() {
    r.resetStepLine()
    r.mlStep = []map[string]string{}
}

func (r *Runner) recover() {
    if rec := recover(); rec != nil {
        if rec == "Pending" {
            r.scenarioIsPending = true
        } else {
            panic(rec)
        }
    }
}

func (r *Runner) executeStepDef(currStep step) {
    defer r.recover()
    for _, step := range r.steps {
        if step.execute(currStep, r.mlStep) {
            r.StepCount++
            return
        }
    }
}

func (r *Runner) executeFirstMatchingStep() {
    currStep := r.currStepLine()

    defer r.reset()
    if r.hasOutstandingStep() {
        r.executeStepDef(currStep)
    }
}

func (r *Runner) callSetUp() {
    if r.setUp != nil {
        r.setUp()
    }
}

func (r *Runner) callTearDown() {
    if r.tearDown != nil {
        r.tearDown()
    }
}

func (r *Runner) parseAsStep(line string) (bool, string) {
    givenMatch, _ := re.Compile(`(Given|When|Then|And|But|\*) (.*?)\s*$`)
    if s := givenMatch.FindStringSubmatch(line); s != nil && len(s) > 1 {
        return true, s[2]
    }
    return false, ""
}

func (r *Runner) isScenarioLine(line string) (bool) {
    scenarioMatch, _ := re.Compile(`Scenario:\s*(.*?)\s*$`)
    if s := scenarioMatch.FindStringSubmatch(line); s != nil {
        return true
    }
    return false
}

func (r *Runner) isFeatureLine(line string) bool {
    featureMatch, _ := re.Compile(`Feature:\s*(.*?)\s*$`)
    if s := featureMatch.FindStringSubmatch(line); s != nil {
        return true
    }
    return false
}

func parseTableLine(line string) (fields []string) {
    mlMatch, _ := re.Compile(`^\s*\|.*\|\s*$`)
    if mlMatch.MatchString(line) {
        tmpFields := strings.Split(line, "|")
        fields = tmpFields[1:len(tmpFields)-1]
        for i, f := range fields {
            fields[i] = strings.TrimSpace(f)
        }
    }
    return
}

func createTableMap(keys []string, fields []string) (l map[string]string) {
    l = map[string]string{}
    for i, k := range keys {
        l[k] = fields[i]
    }
    return
}

func (r *Runner) isBackgroundLine(line string) bool {
    backgroundMatch, _ := re.Compile(`Background:`)
    if s := backgroundMatch.FindStringSubmatch(line); s != nil {
        return true
    }
    return false
}

func (r *Runner) executeStep(line string) {
    if r.collectBackground {
        r.background = append(r.background, StepFromString(line))
    } else {
        r.addStepLine(line)
    }
}

func (r *Runner) startScenario() {
    r.callTearDown()
    r.collectBackground = false
    r.scenarioIsPending = false
    r.currScenario = scenario{}
    r.scenarios = append(r.scenarios, r.currScenario)
    r.callSetUp()
    for _, bline := range r.background {
        r.executeStepDef(bline)
    }
}

func (r *Runner) executeLastScenario() {
    for _, line := range r.currScenario {
        if !r.scenarioIsPending {
            r.executeStepDef(line)
        }
    }
}

func (r *Runner) setMlKeys(data []string) {
    r.currScenario[len(r.currScenario)-1].setMlKeys(data)
}

func (r *Runner) addMlStep(data map[string]string) {
    r.mlStep = append(r.mlStep, data)
    r.currScenario[len(r.currScenario)-1].addMlData(data)
}

func (r *Runner) step(line string) {
    fields := parseTableLine(line)
    if isStep, data := r.parseAsStep(line); isStep {
        r.executeFirstMatchingStep()
        // If the previous step didn't make us pending, go ahead and execute the new one when appropriate
        if !r.scenarioIsPending {
            r.executeStep(data)
        }
    } else if r.isScenarioLine(line) {
        r.executeFirstMatchingStep()
        r.startScenario()
    } else if r.isFeatureLine(line) {
        // Do Nothing!
    } else if r.isBackgroundLine(line) {
        r.collectBackground = true
    } else if len(fields) > 0 {
        s := r.currScenario[len(r.currScenario)-1]
        if len(s.keys) == 0 {
            r.setMlKeys(fields)
        } else if len(fields) != len(s.keys) {
            panic(fmt.Sprintf("Wrong number of fields in multi-line step [%v] - expected %d fields but found %d", line, len(s.keys), len(fields)))
        } else if len(fields) > 0 {
            l := createTableMap(s.keys, fields)
            r.addMlStep(l)
        }
    }
}

// Once the step definitions are Register()'d, use Execute() to
// parse and execute Gherkin data.
func (r *Runner) Execute(file string) {
    lines := strings.Split(file, "\n")
    for _, line := range lines {
        r.step(line)
    }
    r.executeFirstMatchingStep()
    r.callTearDown()
}

// Use this function to let the user know that this
// test is not complete.
func Pending() {
    panic("Pending")
}
