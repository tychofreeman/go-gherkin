package gherkin

import (
    re "regexp"
    "io"
)

type stepdef struct {
    r *re.Regexp
    f func(*World)
}

func createstepdef(p string, f func(*World)) stepdef {
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

func (s stepdef) String() string {
    return s.r.String()
}
