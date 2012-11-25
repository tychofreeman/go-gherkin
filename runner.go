package gherkin

import (
    re "regexp"
    "strings"
    "fmt"
    "io"
    "io/ioutil"
    "path/filepath"
    "os"
)

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
    r.steps = append(r.steps, createstepdef(pattern, f))
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

// Once the step definitions are Register()'d, use Run() to
// locate all *.feature files within the feature/ subdirectory
// of the current directory.
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

// By default, Runner uses os.Stdout to write to. However, it may be useful
// to redirect. To do so, provide an io.Writer here.
func (r *Runner) SetOutput(w io.Writer) {
    r.output = w
}
