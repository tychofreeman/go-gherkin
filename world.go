package gherkin

import (
    "fmt"
    "io"
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

// Allows World to be used with the go-matchers AssertThat() function.
func (w *World) Errorf(format string, args ...interface{}) {
    w.gotAnError = true
    if w.output != nil {
        fmt.Fprintf(w.output, format, args)
    }
}
