package gherkin

import (
    "bytes"
    "fmt"
)

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

func (s *step) recoverPending() {
    if rec := recover(); rec != nil {
        if rec == "Pending" {
            s.isPending = true
        } else {
            panic(rec)
        }
    }
}

func (currStep *step) executeStepDef(steps []stepdef) bool {
    defer currStep.recoverPending()
    for _, stepd := range steps {
            //fmt.Printf("Executing step %s with stepdef %d (%v)\n", currStep, i, stepd)
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
