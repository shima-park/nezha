// Package inject provides utilities for mapping and injecting dependencies in various ways.
package inject

import (
	"fmt"
	"reflect"
	"sync"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

// Injector represents an interface for mapping and injecting dependencies into structs
// and function arguments.
type Injector interface {
	Applicator
	Invoker
	TypeMapper
	// SetParent sets the parent of the injector. If the injector cannot find a
	// dependency in its Type map it will check its parent before returning an
	// error.
	SetParent(Injector)
}

// Applicator represents an interface for mapping dependencies to a struct.
type Applicator interface {
	// Maps dependencies in the Type map to each field in the struct
	// that is tagged with 'inject'. Returns an error if the injection
	// fails.
	Apply(interface{}) error
}

// Invoker represents an interface for calling functions via reflection.
type Invoker interface {
	// Invoke attempts to call the interface{} provided as a function,
	// providing dependencies for function arguments based on Type. Returns
	// a slice of reflect.Value representing the returned values of the function.
	// Returns an error if the injection fails.
	Invoke(interface{}) ([]reflect.Value, error)
}

// TypeMapper represents an interface for mapping interface{} values based on type.
type TypeMapper interface {
	// Maps the interface{} value based on its immediate type from reflect.TypeOf.
	Map(value interface{}, name string) TypeMapper
	// Maps the interface{} value based on the pointer of an Interface provided.
	// This is really only useful for mapping a value as an interface, as interfaces
	// cannot at this time be referenced directly without a pointer.
	MapTo(value interface{}, name string, ifacePtr interface{}) TypeMapper
	// Provides a possibility to directly insert a mapping based on type and value.
	// This makes it possible to directly map type arguments not possible to instantiate
	// with reflect like unidirectional channels.
	Set(typ reflect.Type, name string, value reflect.Value) TypeMapper
	// Returns the Value that is mapped to the current type. Returns a zeroed Value if
	// the Type has not been mapped.
	Get(typ reflect.Type, name string) reflect.Value

	MapValues(vals ...reflect.Value) error
}

type injector struct {
	lock   sync.RWMutex
	values map[reflect.Type]map[string]reflect.Value
	parent Injector
}

// InterfaceOf dereferences a pointer to an Interface type.
// It panics if value is not an pointer to an interface.
func InterfaceOf(value interface{}) reflect.Type {
	t := reflect.TypeOf(value)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Interface {
		panic("Called inject.InterfaceOf with a value that is not a pointer to an interface. (*MyInterface)(nil)")
	}

	return t
}

// New returns a new Injector.
func New() Injector {
	return &injector{
		values: make(map[reflect.Type]map[string]reflect.Value),
	}
}

// Invoke attempts to call the interface{} provided as a function,
// providing dependencies for function arguments based on Type.
// Returns a slice of reflect.Value representing the returned values of the function.
// Returns an error if the injection fails.
// It panics if f is not a function
func (inj *injector) Invoke(f interface{}) ([]reflect.Value, error) {
	t := reflect.TypeOf(f)

	var in = make([]reflect.Value, t.NumIn()) //Panic if t is not kind of Func
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)

		for argType.Kind() == reflect.Ptr {
			argType = argType.Elem()
		}

		if argType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("Cannot support types other than structures %v", argType)
		}

		val := reflect.New(argType)

		if err := inj.Apply(val.Interface()); err != nil {
			return nil, err
		}

		if t.In(i).Kind() == reflect.Struct { // 请求参数需要struct而不是ptr准换一下
			val = val.Elem()
		}

		in[i] = val
	}

	return reflect.ValueOf(f).Call(in), nil
}

// Maps dependencies in the Type map to each field in the struct
// that is tagged with 'inject'.
// Returns an error if the injection fails.
func (inj *injector) Apply(val interface{}) error {
	v := reflect.ValueOf(val)

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil // Should not panic here ?
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		structField := t.Field(i)
		if f.CanSet() && (structField.Tag == "inject" || structField.Tag.Get("inject") != "") {
			name := structField.Tag.Get("inject")
			if name == "" {
				name = structField.Name
			}

			ft := f.Type()
			v := inj.Get(ft, name)
			if !v.IsValid() {
				return fmt.Errorf("Value not found for type: %v name: %v", ft, name)
			}

			f.Set(v)
		}

	}

	return nil
}

// Maps the concrete value of val to its dynamic type using reflect.TypeOf,
// It returns the TypeMapper registered in.
func (i *injector) Map(val interface{}, name string) TypeMapper {
	return i.Set(reflect.TypeOf(val), name, reflect.ValueOf(val))
}

func (i *injector) MapTo(val interface{}, name string, ifacePtr interface{}) TypeMapper {
	return i.Set(InterfaceOf(ifacePtr), name, reflect.ValueOf(val))
}

// Maps the given reflect.Type to the given reflect.Value and returns
// the Typemapper the mapping has been registered in.
func (i *injector) Set(typ reflect.Type, name string, val reflect.Value) TypeMapper {
	return i.set(typ, name, val)
}

func (i *injector) get(typ reflect.Type, name string) (val reflect.Value) {
	i.lock.RLock()
	defer i.lock.RUnlock()

	if m, ok := i.values[typ]; ok {
		val = m[name]
		return
	}
	return
}

func (i *injector) set(typ reflect.Type, name string, val reflect.Value) TypeMapper {
	i.lock.Lock()
	defer i.lock.Unlock()

	var m map[string]reflect.Value
	var ok bool
	if m, ok = i.values[typ]; !ok {
		m = map[string]reflect.Value{}
		i.values[typ] = m
	}

	m[name] = val
	return i
}

func (i *injector) Get(t reflect.Type, name string) reflect.Value {
	val := i.get(t, name)
	if val.IsValid() {
		return val
	}

	// no concrete types found, try to find implementors
	// if t is an interface
	if t.Kind() == reflect.Interface {
		i.lock.RLock()
		for k, m := range i.values {
			for n, v := range m {
				if n == name && k.Implements(t) {
					val = v
					i.lock.RUnlock()
					break
				}
			}
		}
		i.lock.RUnlock()
	}

	// Still no type found, try to look it up on the parent
	if !val.IsValid() && i.parent != nil {
		val = i.parent.Get(t, name)
	}

	return val
}

func (i *injector) SetParent(parent Injector) {
	i.parent = parent
}

func (i *injector) MapValues(vals ...reflect.Value) error {
	for _, val := range vals {
		// 处理返回值中带error的情况
		if val.Type().Implements(errorInterface) && !val.IsNil() {
			return val.Interface().(error)
		}

		// 获取返回值中struct的部分
		for val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() != reflect.Struct {
			continue
		}

		typ := val.Type()

		for k := 0; k < val.NumField(); k++ {
			f := val.Field(k)
			structField := typ.Field(k)
			if f.IsValid() && (structField.Tag == "inject" || structField.Tag.Get("inject") != "") {
				name := structField.Tag.Get("inject")
				if name == "" {
					name = structField.Name
				}
				if f.Type().Kind() == reflect.Interface {
					nilPtr := reflect.New(f.Type())
					i.MapTo(f.Interface(), name, nilPtr.Interface())
				} else {
					i.Set(f.Type(), name, f)
				}
			}
		}
	}
	return nil
}
