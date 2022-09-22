package assert

import (
	"fmt"
	"reflect"
	"testing"
)

type Stub string

type Lister interface {
	List() ([]*Stub, error)
}

func importProcessSTubList(restClient Lister, resourceName string, targetFolder string) {

}

type Assert interface {
	Errorf(format string, args ...any)
	Fail()
	True(v bool)
	Nil(v interface{})
	Equals(expected interface{}, actual interface{})
	Equalsf(expected interface{}, actual interface{}, format string, args ...any)
}

func New(t *testing.T) Assert {
	return &assert{t}
}

type assert struct {
	t *testing.T
}

func (a *assert) Errorf(format string, args ...any) {
	a.t.Helper()
	a.t.Errorf(format, args...)
}

func (a *assert) Fail() {
	a.t.Helper()
	a.t.Fail()
}

func (a *assert) True(v bool) {
	a.t.Helper()
	if !v {
		a.Fail()
	}
}

func (a *assert) Nil(v interface{}) {
	a.t.Helper()
	if v != nil {
		a.Fail()
	}
}

func (a *assert) Equals(expected interface{}, actual interface{}) {
	a.t.Helper()
	if res := equals(expected, actual); res != "" {
		a.Errorf("%s", res)
	}
}

func (a *assert) Equalsf(expected interface{}, actual interface{}, format string, args ...any) {
	a.t.Helper()
	if res := equals(expected, actual); res != "" {
		a.Errorf("%s: %s", fmt.Sprintf(format, args...), res)
	}
}

func equals(expected interface{}, actual interface{}) string {
	if expected == nil {
		if actual == nil {
			return ""
		}
		return fmt.Sprintf("expected: nil, actual: %v", actual)
	} else if actual == nil {
		return fmt.Sprintf("expected: %v, actual: nil", expected)
	}
	if reflect.TypeOf(expected) != reflect.TypeOf(actual) {
		return fmt.Sprintf("expected: %v (type %T), actual: %v (type %T)", expected, expected, actual, actual)
	}
	switch et := expected.(type) {
	case map[string]interface{}:
		return mapEquals(et, actual.(map[string]interface{}))
	}
	tExpected := reflect.TypeOf(expected)
	switch kind := tExpected.Kind(); kind {
	case reflect.Slice, reflect.Array:
		return sliceEquals(reflect.ValueOf(expected), reflect.ValueOf(actual))
	case reflect.Map:
		return reflectMapEquals(expected, actual)
	default:
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Sprintf("expected: %v, actual: %v", expected, actual)
		}
	}
	return ""
}

func isNil(v reflect.Value) bool {
	switch kind := v.Kind(); kind {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return !v.IsValid() || v.IsNil()
	default:
		return false
	}
}

func sliceEquals(vExpected reflect.Value, vActual reflect.Value) string {
	lenExpected := vExpected.Len()
	lenActual := vActual.Len()
	if lenExpected != lenActual {
		return fmt.Sprintf("slice/array lengths don't match - expected: %d, actual: %d", lenExpected, lenActual)
	}
	for idx := 0; idx < vExpected.Len(); idx++ {
		elemExpected := vExpected.Index(idx)
		elemActual := vActual.Index(idx)
		if isNil(elemExpected) {
			if !isNil(elemActual) {
				return fmt.Sprintf("[%d] expected: nil, actual: %v", idx, elemActual.Interface())
			}
		} else if isNil(elemActual) {
			return fmt.Sprintf("[%d] expected: %v, actual: nil", idx, elemExpected.Interface())
		}
		if res := equals(elemExpected.Interface(), elemActual.Interface()); res != "" {
			return fmt.Sprintf("[%d] expected: %v, actual: %v", idx, elemExpected.Interface(), elemActual.Interface())
		}
	}
	return ""
}

func reflectMapEquals(expected interface{}, actual interface{}) string {
	if expected == nil {
		if actual == nil {
			return ""
		}
		return "should be nil"
	} else if actual == nil {
		return "should not be nil"
	}
	vExpected := reflect.ValueOf(expected)
	vActual := reflect.ValueOf(actual)
	for _, k := range vExpected.MapKeys() {
		ve := vExpected.MapIndex(k)
		va := vActual.MapIndex(k)
		if !va.IsZero() {
			if res := equals(ve.Interface(), va.Interface()); res != "" {
				return fmt.Sprintf("[\"%v\"] %s", k.Interface(), res)
			}
		} else {
			return fmt.Sprintf("[\"%v\"] not found>", k.Interface())
		}
	}
	for _, k := range vActual.MapKeys() {
		ve := vExpected.MapIndex(k)
		if ve.IsZero() {
			return fmt.Sprintf("[\"%v\"] shouldn't exist", k.Interface())
		}
	}
	return ""
}

func mapEquals(expected map[string]interface{}, actual map[string]interface{}) string {
	if expected == nil {
		if actual == nil {
			return ""
		}
		return fmt.Sprintf("expected: nil, actual: %v", actual)
	} else if actual == nil {
		return fmt.Sprintf("expected: %v, actual: nil", expected)
	}
	for k, ve := range expected {
		if va, ok := actual[k]; ok {
			if res := equals(ve, va); res != "" {
				return fmt.Sprintf("[\"%s\"] %s", k, res)
			}
		} else {
			return fmt.Sprintf("[\"%s\"] - expected: %v, actual: <notfound>", k, ve)
		}
	}
	for k, va := range actual {
		if _, ok := expected[k]; !ok {
			return fmt.Sprintf("[\"%s\"] shouldn't exist, actual: %v", k, va)
		}
	}
	return ""
}
