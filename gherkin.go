// Support the Gherkin language, as found in Ruby's Cucumber and Python's Lettuce projects.
package gherkin

import (
    re "regexp"
    "strings"
    "fmt"
    "io"
    "io/ioutil"
    "path/filepath"
    "os"
//    "runtime/debug"
    "bytes"
)

// Passed to each step-definition
type World struct {
    regexParams []string
    regexParamIndex int
    multiStep []map[string]string
    output io.Writer
    gotAnError bool
}

// Allows access to step definition regular expression captures.
func (w World) GetRegexParam() string {
    w.regexParamIndex++
    if w.regexParamIndex >= len(w.regexParams) {
        panic("GetRegexParam() called too many times.")
    }
    return w.regexParams[w.regexParamIndex]
}

func (w *World) Errorf(format string, args ...interface{}) {
    w.gotAnError = true
    if w.output != nil {
        fmt.Fprintf(w.output, format, args)
    }
}

type stepdef struct {
    r *re.Regexp
    f func(*World)
}

func createstep(p string, f func(*World)) stepdef {
    r, _ := re.Compile(p)
    return stepdef{r, f}
}

func (s stepdef) execute(line *step, output io.Writer) bool {
    if s.r.MatchString(line.String()) {
        if s.f != nil {
            substrs := s.r.FindStringSubmatch(line.String())
            w := &World{regexParams:substrs, multiStep:line.mldata, output: output} 
            defer func() { line.hasErrors = w.gotAnError }()
            s.f(w)
        }
        return true
    }
    return false
}

type step struct {
    line string
    orig string
    keys []string
    mldata []map[string]string
    isPending bool
    errors bytes.Buffer
    hasErrors bool
}

func (s step) String() string {
    return s.line
}

func StepFromString(in string) step{
    return step{ line : in, keys: []string{}, mldata : []map[string]string{} }
}

func StepFromStringAndOrig(in, orig string) step{
    return step{ line : in, orig: orig, keys: []string{}, mldata : []map[string]string{} }
}

func (s *step) addMlData(line map[string]string) {
    s.mldata = append(s.mldata, line)
}

func (s *step) recover() {
    if rec := recover(); rec != nil {
        if rec == "Pending" {
            s.isPending = true
        } else {
            panic(rec)
        }
    }
}

func (currStep *step) executeStepDef(steps []stepdef) bool {
    for _, stepd := range steps {
        defer currStep.recover()
        if stepd.execute(currStep, &currStep.errors) {
            return true
        }
    }
    fmt.Fprintf(&currStep.errors, `Could not find step definition for "%s"` + "\n", currStep.orig)
    return false
}

func (s *step) setMlKeys(keys []string) {
    s.keys = keys
}

type scenario_outline struct {
    steps []step
    keys []string
    isPending bool
}

func ScenarioOutline() scenario_outline {
    return scenario_outline{}
}

func (so *scenario_outline) AddStep(s step) {
    so.steps = append(so.steps, s)
}

func (so scenario_outline) CreateForExample(example map[string]string) scenario {
    s := scenario{}
    for _, currStep := range so.steps {
        l := currStep.line

        for k, v := range example {
            r, _ := re.Compile("<" + k + ">")
            l = r.ReplaceAllString(l, v)
        }
        s.steps = append(s.steps, StepFromString(l))
    }

    return s
}

func (so *scenario_outline) Last() *step {
    if so.steps != nil {
        return &so.steps[len(so.steps)-1]
    }
    return nil
}

func (so *scenario_outline) Execute(s []stepdef, output io.Writer) {
}

type Scenario interface {
    AddStep(step)
    Last() *step
    Execute([]stepdef, io.Writer)
}

type scenario struct {
    steps []step
    isPending bool
    orig string
}

func (scen *scenario) AddStep(stp step) {
    if scen.steps == nil {
        scen.steps = []step{stp}
    } else {
        scen.steps = append(scen.steps, stp)
    }
}

func (s *scenario) Last() *step {
    if len(s.steps) > 0 {
        return &s.steps[len(s.steps)-1]
    }
    return nil
}

func (s *scenario) Execute(stepdefs []stepdef, output io.Writer) {
    if output != nil {
        fmt.Fprintf(output, s.orig + "\n")
    }
    isPending := false
    for _, line := range s.steps {
        if !isPending {
            line.executeStepDef(stepdefs)
        }
        if line.isPending {
            if output != nil {
                fmt.Fprintf(output, "PENDING - %s\n", line.orig)
            }
            isPending = true
        } else if isPending {
            if output != nil {
                fmt.Fprintf(output, "Skipped - %s\n", line.orig)
            }
        } else {
            if output != nil {
                fmt.Fprintf(output, "        - %s\n", line.orig)
            }
        }
        if line.hasErrors {
            // The '&' is necessary to make .errors conform to the io.Writer interface
            fmt.Fprintf(output, "%v", &line.errors)
        }
    }
}

type Runner struct {
    steps []stepdef
    background Scenario
    isExample bool
    setUp func()
    tearDown func()
    currScenario Scenario
    scenarios []Scenario
    output io.Writer
}

func (r *Runner) addStepLine(line, orig string) {
    r.currScenario.AddStep(StepFromStringAndOrig(line, orig))
}

func (r *Runner) currStepLine() step {
    l := r.currStep()
    if l == nil {
        return StepFromString("")
    }
    return *l
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
    s := []Scenario{&scenario{}}
    return &Runner{[]stepdef{}, nil, false, nil, nil, nil, s, os.Stdout}
}

func createWriterlessRunner() *Runner {
    r := CreateRunner()
    r.output = nil
    return r
}

// Register a step definition. This requires a regular expression
// pattern and a function to execute.
func (r *Runner) RegisterStepDef(pattern string, f func(*World)) {
    r.steps = append(r.steps, createstep(pattern, f))
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

func (r *Runner) runBackground() {
    if r.background != nil {
        r.background.Execute(r.steps, r.output)
    }
}

func parseAsStep(line string) (bool, string) {
    givenMatch, _ := re.Compile(`^\s*(Given|When|Then|And|But|\*)\s+(.*?)\s*$`)
    if s := givenMatch.FindStringSubmatch(line); s != nil && len(s) > 1 {
        return true, s[2]
    }
    return false, ""
}

func isScenarioOutline(line string) bool {
    return lineMatches(`^\s*Scenario Outline:\s*(.*?)\s*$`, line)
}

func isExampleLine(line string) bool {
    return lineMatches(`^\s*Examples:\s*(.*?)\s*$`, line)
}

func isScenarioLine(line string) (bool) {
    return lineMatches(`^\s*Scenario:\s*(.*?)\s*$`, line)
}

func isFeatureLine(line string) bool {
    return lineMatches(`Feature:\s*(.*?)\s*$`, line)
}
func isBackgroundLine(line string) bool {
    return lineMatches(`^\s*Background:`, line)
}

func lineMatches(spec, line string) bool {
    featureMatch, _ := re.Compile(spec)
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

func (r *Runner) resetWithScenario(s Scenario) {
    r.isExample = false
    r.scenarios = append(r.scenarios, s)
    r.currScenario = r.scenarios[len(r.scenarios)-1]
}

func (r *Runner) startScenarioOutline() {
    r.resetWithScenario(&scenario_outline{})
}

func (r *Runner) startScenario(orig string) {
    r.resetWithScenario(&scenario{orig: orig})
}

func (r *Runner) currStep() *step {
    if r.currScenario != nil {
        return r.currScenario.Last()
    }
    return nil
}


func (r *Runner) setMlKeys(data []string) {
    r.currStep().setMlKeys(data)
}

func (r *Runner) addMlStep(data map[string]string) {
    r.currStep().addMlData(data)
}

func (r *Runner) step(line string) {
    fields := parseTableLine(line)
    isStep, data := parseAsStep(line)
    if r.currScenario != nil && isStep {
        r.addStepLine(data, line)
    } else if isScenarioOutline(line) {
        r.startScenarioOutline()
    } else if isScenarioLine(line) {
        r.startScenario(line)
    } else if isFeatureLine(line) {
        // Do Nothing!
    } else if isBackgroundLine(line) {
        r.startScenario(line)
        r.background = r.currScenario
    } else if isExampleLine(line) {
        r.isExample = true
    } else if r.isExample && len(fields) > 0 {
        switch scen := r.currScenario.(type) {
            case *scenario_outline:
                if scen.keys == nil {
                    scen.keys = fields
                } else {
                    newScenario := scen.CreateForExample(createTableMap(scen.keys, fields))
                    r.scenarios = append(r.scenarios, &newScenario)
                }
            default:
        }
    } else if r.currStep() != nil && len(fields) > 0 {
        s := *r.currStep()
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

func (r *Runner) executeScenario(scenario Scenario) {
    r.callSetUp()
    r.runBackground()
    scenario.Execute(r.steps, r.output)
    r.callTearDown()
}

// Once the step definitions are Register()'d, use Execute() to
// parse and execute Gherkin data.
func (r *Runner) Execute(file string) {
    lines := strings.Split(file, "\n")
    for _, line := range lines {
        r.step(line)
    }
    for _, scenario := range r.scenarios {
        r.executeScenario(scenario)
    }
}

func (r *Runner) Run() {
    featureMatch, _ := re.Compile(`.*\.feature`)
    filepath.Walk("features", func(walkPath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.Name() != "features" && info.IsDir() {
            return filepath.SkipDir
        } else if !info.IsDir() && featureMatch.MatchString(info.Name()) {
            file, _ := os.Open(walkPath)
            data, _ := ioutil.ReadAll(file)
            r.Execute(string(data))
        }
        return nil
    })
}

func (r *Runner) SetOutput(w io.Writer) {
    r.output = w
}

// Use this function to let the user know that this
// test is not complete.
func Pending() {
    panic("Pending")
}
