package goink

import (
	"regexp"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/pkg/errors"
)

var regReplaceDot = regexp.MustCompile(`\.(\w+)`)

// exprc in the line
type exprc struct {
	// env     map[string]interface{}
	program    *vm.Program
	programInt *vm.Program
	programNil *vm.Program
	raw        string
}

// newExprc creates a condition with the given expr
func newExprc(code string) (*exprc, error) {
	cond := &exprc{raw: code}
	c := regReplaceDot.ReplaceAllString(code, PathSplit+"$1")

	program, err := expr.Compile(c, expr.Env(nil))

	if err != nil {
		return nil, err
	}

	cond.program = program

	c2 := regReplaceDot.ReplaceAllString(code, PathSplit+"$1")
	c2 = strings.ReplaceAll(c2, "not ", "0 == ")
	programInt, _ := expr.Compile(c2, expr.Env(nil))

	c2 = regReplaceDot.ReplaceAllString(code, PathSplit+"$1")
	c2 = strings.ReplaceAll(c2, "not ", "nil == ")
	programNil, _ := expr.Compile(c2, expr.Env(nil))

	cond.programInt = programInt
	cond.programNil = programNil

	return cond, nil
}

// Bool return the exprc result as bool value
func (c *exprc) Bool(count map[string]interface{}) (bool, error) {

	output, err := expr.Run(c.program, count)
	if err != nil {
		if strings.Contains(err.Error(), "int, not bool") {
			// newCount := make(map[string]interface{}, len(count))
			// for k, v := range count {
			// 	vAsInt, wasInt := v.(int)
			// 	if wasInt {
			// 		if vAsInt == 0 {
			// 			newCount[k] = false
			// 		} else {
			// 			newCount[k] = true
			// 		}
			// 	} else {
			// 		newCount[k] = v
			// 	}

			// }

			//So basically, some of our stuff (like how many times we've visited a knot) are stored as Integers (i.e. 1) instead of boolean values.
			//So we have to run a different program to check for that stuff since the Golang stuff is type-checked (Javascript would just have let this go and treated)
			//any non-zero-value as "truthy".
			output, err = expr.Run(c.programInt, count)
			if err != nil {
				return false, err
			}
		} else if strings.Contains(err.Error(), "nil, not bool") {
			output, err = expr.Run(c.programNil, count)
			if err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	}

	// fmt.Println(c.program.Source.Content(), output, count["Knot_A-gather"])

	b, ok := output.(bool)
	if ok {
		return b, nil
	}

	i, ok := output.(int)
	if ok {
		return (i > 0), nil
	}

	if output == nil {
		return false, nil
	}

	return false, errors.Errorf("output is not a bool value: %v", output)
}
