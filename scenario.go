package gherkin

import (
    "fmt"
    "io"
    re "regexp"
)

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
        if output != nil {
            fmt.Fprintf(output, "%v", &line.errors)
        }
    }
}
