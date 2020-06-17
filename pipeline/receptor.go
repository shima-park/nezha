package pipeline

import (
	"reflect"

	"github.com/shima-park/nezha/common/inject"
	"github.com/shima-park/nezha/processor"
)

type Receptor struct {
	StructName      string
	StructFieldName string
	InjectName      string
	ReflectType     string
}

func getFuncReqAndRespReceptorList(f interface{}) ([]Receptor, []Receptor) {
	if err := processor.Validate(f); err != nil {
		return nil, nil
	}

	t := reflect.TypeOf(f)

	var reqReceptors []Receptor
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)

		for argType.Kind() == reflect.Ptr {
			argType = argType.Elem()
		}

		if argType.Kind() != reflect.Struct {
			continue
		}

		val := reflect.New(argType)

		for val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() != reflect.Struct {
			continue
		}

		typ := val.Type()

		for i := 0; i < val.NumField(); i++ {
			f := val.Field(i)
			structField := typ.Field(i)
			injectName := structField.Tag.Get("inject")

			var tt reflect.Type
			if f.Type().Kind() == reflect.Interface {
				nilPtr := reflect.New(f.Type())
				tt = inject.InterfaceOf(nilPtr.Interface())
			} else {
				tt = f.Type()
			}

			reqReceptors = append(reqReceptors, Receptor{
				StructName:      typ.Name(),
				StructFieldName: structField.Name,
				InjectName:      injectName,
				ReflectType:     tt.String(),
			})
		}
	}

	var respReceptors []Receptor
	for i := 0; i < t.NumOut(); i++ {
		outType := t.Out(i)

		if outType.Implements(errorInterface) {
			continue
		}

		for outType.Kind() == reflect.Ptr {
			outType = outType.Elem()
		}

		if outType.Kind() != reflect.Struct {
			continue
		}

		val := reflect.New(outType)
		for val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() != reflect.Struct {
			continue
		}

		typ := val.Type()

		for i := 0; i < val.NumField(); i++ {
			f := val.Field(i)
			structField := typ.Field(i)
			injectName := structField.Tag.Get("inject")

			var tt reflect.Type
			if f.Type().Kind() == reflect.Interface {
				nilPtr := reflect.New(f.Type())
				tt = inject.InterfaceOf(nilPtr.Interface())
			} else {
				tt = f.Type()
			}

			respReceptors = append(respReceptors, Receptor{
				StructName:      typ.Name(),
				StructFieldName: structField.Name,
				InjectName:      injectName,
				ReflectType:     tt.String(),
			})
		}
	}
	return reqReceptors, respReceptors
}
