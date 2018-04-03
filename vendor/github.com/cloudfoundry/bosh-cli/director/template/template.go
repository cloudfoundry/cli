package template

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
)

type Template struct {
	bytes []byte
}

type EvaluateOpts struct {
	ExpectAllKeys         bool
	ExpectAllVarsUsed     bool
	PostVarSubstitutionOp patch.Op
	UnescapedMultiline    bool
}

func NewTemplate(bytes []byte) Template {
	return Template{bytes: bytes}
}

func (t Template) Evaluate(vars Variables, op patch.Op, opts EvaluateOpts) ([]byte, error) {
	var obj interface{}

	err := yaml.Unmarshal(t.bytes, &obj)
	if err != nil {
		return []byte{}, err
	}

	if op != nil {
		obj, err = op.Apply(obj)
		if err != nil {
			return []byte{}, err
		}
	}

	obj, err = t.interpolateRoot(obj, newVarsTracker(vars, opts.ExpectAllKeys, opts.ExpectAllVarsUsed))
	if err != nil {
		return []byte{}, err
	}

	if opts.PostVarSubstitutionOp != nil {
		obj, err = opts.PostVarSubstitutionOp.Apply(obj)
		if err != nil {
			return []byte{}, err
		}
	}

	if _, ok := obj.(string); opts.UnescapedMultiline && ok {
		return []byte(fmt.Sprintf("%s\n", obj)), nil
	}

	bytes, err := yaml.Marshal(obj)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func (t Template) interpolateRoot(obj interface{}, tracker varsTracker) (interface{}, error) {
	err := tracker.ExtractDefinitions(obj)
	if err != nil {
		return nil, err
	}

	obj, err = interpolator{}.Interpolate(obj, varsLookup{tracker})
	if err != nil {
		return nil, err
	}

	obj, err = tracker.FoundVars(obj)
	if err != nil {
		return nil, err
	}

	return obj, tracker.Error()
}

type interpolator struct{}

var (
	interpolationRegex         = regexp.MustCompile(`\(\((!?[-/\.\w\pL]+)\)\)`)
	interpolationAnchoredRegex = regexp.MustCompile("\\A" + interpolationRegex.String() + "\\z")
)

func (i interpolator) Interpolate(node interface{}, varsLookup varsLookup) (interface{}, error) {
	switch typedNode := node.(type) {
	case map[interface{}]interface{}:
		for k, v := range typedNode {
			evaluatedValue, err := i.Interpolate(v, varsLookup)
			if err != nil {
				return nil, err
			}

			evaluatedKey, err := i.Interpolate(k, varsLookup)
			if err != nil {
				return nil, err
			}

			delete(typedNode, k) // delete in case key has changed
			typedNode[evaluatedKey] = evaluatedValue
		}

	case []interface{}:
		for idx, x := range typedNode {
			var err error
			typedNode[idx], err = i.Interpolate(x, varsLookup)
			if err != nil {
				return nil, err
			}
		}

	case string:
		for _, name := range i.extractVarNames(typedNode) {
			foundVal, found, err := varsLookup.Get(name)
			if err != nil {
				return nil, bosherr.WrapErrorf(err, "Finding variable '%s'", name)
			}

			if found {
				// ensure that value type is preserved when replacing the entire field
				if interpolationAnchoredRegex.MatchString(typedNode) {
					return foundVal, nil
				}

				switch foundVal.(type) {
				case string, int, int16, int32, int64, uint, uint16, uint32, uint64:
					foundValStr := fmt.Sprintf("%v", foundVal)
					typedNode = strings.Replace(typedNode, fmt.Sprintf("((%s))", name), foundValStr, -1)
					typedNode = strings.Replace(typedNode, fmt.Sprintf("((!%s))", name), foundValStr, -1)
				default:
					errMsg := "Invalid type '%T' for value '%v' and variable '%s'. Supported types for interpolation within a string are integers and strings."
					return nil, fmt.Errorf(errMsg, foundVal, foundVal, name)
				}
			}
		}

		return typedNode, nil
	}

	return node, nil
}

func (i interpolator) extractVarNames(value string) []string {
	var names []string

	for _, match := range interpolationRegex.FindAllSubmatch([]byte(value), -1) {
		names = append(names, strings.TrimPrefix(string(match[1]), "!"))
	}

	return names
}

type varsLookup struct {
	varsTracker
}

func (l varsLookup) Get(name string) (interface{}, bool, error) {
	splitName := strings.Split(name, ".")

	val, found, err := l.varsTracker.Get(splitName[0])
	if !found || err != nil {
		return val, found, err
	}

	if len(splitName) > 1 {
		tokens := []patch.Token{patch.RootToken{}}

		for _, token := range splitName[1:] {
			tokens = append(tokens, patch.KeyToken{Key: token})
		}

		findOp := patch.FindOp{
			Path: patch.NewPointer(tokens),
		}

		val, err = findOp.Apply(val)
		if err != nil {
			return nil, false, err
		}
	}

	return val, true, err
}

type varsTracker struct {
	vars Variables
	defs varDefinitions

	expectAllFound bool
	expectAllUsed  bool

	missing    map[string]struct{} // track missing var names
	visited    map[string]struct{}
	visitedAll map[string]struct{} // track all var names that were accessed
}

func newVarsTracker(vars Variables, expectAllFound, expectAllUsed bool) varsTracker {
	return varsTracker{
		vars:           vars,
		expectAllFound: expectAllFound,
		expectAllUsed:  expectAllUsed,
		missing:        map[string]struct{}{},
		visited:        map[string]struct{}{},
		visitedAll:     map[string]struct{}{},
	}
}

func (t varsTracker) Get(name string) (interface{}, bool, error) {
	t.visitedAll[name] = struct{}{}

	defVarTracker, err := t.scopedVarsTracker(name)
	if err != nil {
		return nil, false, err
	}

	def := t.defs.Find(name)

	def.Options, err = interpolator{}.Interpolate(def.Options, varsLookup{defVarTracker})
	if err != nil {
		return nil, false, bosherr.WrapErrorf(err, "Interpolating variable '%s' definition options", name)
	}

	if len(defVarTracker.missing) > 0 {
		return nil, false, nil
	}

	for name, _ := range defVarTracker.missing {
		t.missing[name] = struct{}{}
	}

	val, found, err := t.vars.Get(def)
	if !found {
		t.missing[name] = struct{}{}
	}

	return val, found, err
}

func (t varsTracker) scopedVarsTracker(name string) (varsTracker, error) {
	if _, found := t.visited[name]; found {
		return varsTracker{}, bosherr.Error("Detected recursion")
	}

	varsTracker := newVarsTracker(t.vars, t.expectAllFound, t.expectAllUsed)
	varsTracker.defs = t.defs
	varsTracker.visited[name] = struct{}{}

	for k, _ := range t.visited {
		varsTracker.visited[k] = struct{}{}
	}

	return varsTracker, nil
}

func (t *varsTracker) ExtractDefinitions(obj interface{}) error {
	if _, isMap := obj.(map[interface{}]interface{}); isMap {
		defsBytes, err := yaml.Marshal(obj)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(defsBytes, &t.defs)
		if err != nil {
			return err
		}
	}

	// Run through all variable definitions in order
	// to provide basic variable dependency management (ie sort definitions manually)
	for _, def := range t.defs.Definitions {
		if len(def.Type) > 0 {
			_, _, err := t.Get(def.Name)
			if err != nil {
				return bosherr.WrapError(err, "Getting all variables from variable definitions sections")
			}
		}
	}

	return nil
}

func (t varsTracker) FoundVars(obj interface{}) (interface{}, error) {
	var foundDefs []string
	var err error

	for _, def := range t.defs.Definitions {
		_, found, err := t.Get(def.Name)
		if err != nil {
			return nil, bosherr.WrapError(err, "Getting all variables from variable definitions sections")
		}
		if found {
			foundDefs = append(foundDefs, def.Name)
		}
	}

	for _, name := range foundDefs {
		removeKeyOp := patch.RemoveOp{
			Path: patch.NewPointer([]patch.Token{
				patch.RootToken{},
				patch.KeyToken{Key: "variables"},
				patch.MatchingIndexToken{Key: "name", Value: name},
			}),
		}

		obj, err = removeKeyOp.Apply(obj)
		if err != nil {
			return nil, err
		}
	}

	return obj, err
}

func (t varsTracker) Error() error {
	var errs []error

	missingErr := t.MissingError()
	if missingErr != nil {
		errs = append(errs, missingErr)
	}

	extraErr := t.ExtraError()
	if extraErr != nil {
		errs = append(errs, extraErr)
	}

	if len(errs) > 0 {
		return bosherr.NewMultiError(errs...)
	}

	return nil
}

func (t varsTracker) MissingError() error {
	if !t.expectAllFound || len(t.missing) == 0 {
		return nil
	}

	return bosherr.WrapError(t.multiErr(t.missing), "Expected to find variables")
}

func (t varsTracker) ExtraError() error {
	if !t.expectAllUsed {
		return nil
	}

	allDefs, err := t.vars.List()
	if err != nil {
		return err
	}

	unusedNames := map[string]struct{}{}

	for _, def := range allDefs {
		if _, found := t.visitedAll[def.Name]; !found {
			unusedNames[def.Name] = struct{}{}
		}
	}

	if len(unusedNames) == 0 {
		return nil
	}

	return bosherr.WrapError(t.multiErr(unusedNames), "Expected to use variables")
}

func (t varsTracker) multiErr(mapWithNames map[string]struct{}) error {
	var names []string
	for name, _ := range mapWithNames {
		names = append(names, name)
	}
	sort.Strings(names)

	var errs []error
	for _, name := range names {
		errs = append(errs, bosherr.Error(name))
	}
	return bosherr.NewMultiError(errs...)
}

type varDefinitions struct {
	Definitions []VariableDefinition `yaml:"variables"`
}

func (defs varDefinitions) Find(name string) VariableDefinition {
	for _, def := range defs.Definitions {
		if def.Name == name {
			return def
		}
	}
	return VariableDefinition{Name: name}
}
