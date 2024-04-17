package dig2

import (
	"reflect"
	"sync"
)

var invalidReflectValue reflect.Value

var vtReflectValue = reflect.TypeOf((*reflect.Value)(nil)).Elem()
var vtErrorType = reflect.TypeOf((*error)(nil)).Elem()

func unwrapTypePtr(v reflect.Type) reflect.Type {
	for {
		k := v.Kind()
		if k == reflect.Ptr {
			v = v.Elem()
		} else {
			return v
		}
	}
}

func unwrapDoubleValueOf(v any) reflect.Value {
	rv := reflect.ValueOf(v)
	for {
		if !rv.IsValid() {
			return rv
		}
		if rv.Type() != vtReflectValue {
			return rv
		}
		if !rv.CanInterface() {
			return rv
		}
		rv = rv.Interface().(reflect.Value)
	}
}

// Returns true if t embeds e or if any of the types embedded by t embed e.
func embedsType(i any, e reflect.Type) bool {
	// TODO: this function doesn't consider e being a pointer.
	// given `type A foo { *In }`, this function would return false for
	// embedding dig.In, which makes for some extra error checking in places
	// that call this funciton. Might be worthwhile to consider reflect.Indirect
	// usage to clean up the callers.

	if i == nil {
		return false
	}

	// maybe it's already a reflect.Type
	t, ok := i.(reflect.Type)
	if !ok {
		// take the type if it's not
		t = reflect.TypeOf(i)
	}

	// We are going to do a breadth-first search of all embedded fields.
	return cacheEmbedsType(t, e)
}

var _cacheEmbedsTypeMap sync.Map

func cacheEmbedsType(t reflect.Type, e reflect.Type) bool {
	type cacheEmbedKey struct {
		t, e reflect.Type
	}
	k := cacheEmbedKey{t: t, e: e}
	if v, has := _cacheEmbedsTypeMap.Load(k); has {
		return v.(bool)
	}
	v := nocacheIsEmbedsType(t, e)
	_cacheEmbedsTypeMap.Store(k, v)
	return v
}

func nocacheIsEmbedsType(t reflect.Type, e reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	if t == e {
		return true
	}
	types := make([]reflect.Type, 0, 1)
	types = append(types, t)
	for pos := 0; pos < len(types); pos++ {
		t = types[pos]
		if t == e {
			return true
		}
		if t.Kind() != reflect.Struct {
			continue
		}
		for i, n := 0, t.NumField(); i < n; i++ {
			f := t.Field(i)
			if f.Anonymous {
				types = append(types, f.Type)
			}
		}
	}

	// If perf is an issue, we can cache known In objects and Out objects in a
	// map[reflect.Type]struct{}.
	return false
}

func isErrorType(t reflect.Type) bool {
	if t == vtErrorType {
		return true
	}
	if cacheIsImplement(t, vtErrorType) {
		return true
	}
	return false
}

var _cacheIsImplementMap sync.Map

func cacheIsImplement(a, b reflect.Type) bool {
	switch a.Kind() {
	case reflect.Interface, reflect.Struct, reflect.Ptr:
	default:
		return false
	}
	switch b.Kind() {
	case reflect.Interface:
	default:
		return false
	}

	type cacheIsImplementKey struct {
		a, b reflect.Type
	}
	k := cacheIsImplementKey{a: a, b: b}
	v, has := _cacheIsImplementMap.Load(k)
	if has {
		return v.(bool)
	}
	var r bool
	r = a.Implements(b)
	_cacheIsImplementMap.Store(k, r)
	return r
}

/*
var _cacheConvertibleToMap sync.Map

func cacheConvertibleTo(a, b reflect.Type) bool {
	type cacheConvertibleToKey struct {
		a, b reflect.Type
	}
	k := cacheConvertibleToKey{a: a, b: b}
	v, has := _cacheConvertibleToMap.Load(k)
	if has {
		return v.(bool)
	}
	var r bool
	r = a.ConvertibleTo(b)
	_cacheConvertibleToMap.Store(k, r)
	return r
}
*/
