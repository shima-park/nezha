package pipeline

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/shima-park/nezha/common/inject"
	"github.com/shima-park/nezha/processor"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

type MissingDependencyError struct {
	Field       string
	ReflectType string
	InjectName  string
}

func (e MissingDependencyError) Error() string {
	return fmt.Sprintf("Value not found for field: %v, type: %v, name: %v",
		e.Field, e.ReflectType, e.InjectName)
}

func check(s *Stream, inj inject.Injector) []error {
	if s == nil || s.processor.Processor == nil {
		return nil
	}

	var errs []error
	if err := processor.Validate(s.processor.Processor); err != nil {
		errs = append(errs, errors.Wrapf(err, "Stream(%s)", s.Name()))
		return errs
	}

	for _, err := range checkDep(inj, s.processor.Processor) {
		errs = append(errs, errors.Wrapf(err, "Stream(%s)", s.Name()))
	}

	for i := 0; i < len(s.childs); i++ {
		for _, err := range check(s.childs[i], inj) {
			errs = append(errs, errors.Wrapf(err, "Stream(%s)", s.Name()))
		}
	}
	return errs
}

func checkDep(inj inject.Injector, f interface{}) []error {
	t := reflect.TypeOf(f)

	var errs []error
	if err := checkIn(inj, t); err != nil {
		errs = append(errs, err...)
	}

	if err := checkOut(inj, t); err != nil {
		errs = append(errs, err...)
	}

	return errs
}

func checkIn(inj inject.Injector, t reflect.Type) []error {
	var errs []error
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)

		for argType.Kind() == reflect.Ptr {
			argType = argType.Elem()
		}

		if argType.Kind() != reflect.Struct {
			errs = append(errs, fmt.Errorf("Cannot support types other than structures %v", argType))
		}

		val := reflect.New(argType)

		for val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() != reflect.Struct {
			continue
		}

		typ := val.Type()
		// 在check过程中没法直接通过injector.Apply来测试是否能注入成功
		// checkout处只能获取到reflect.Type, 对于接口类型的值没法造出reflect.Value
		// 例如：知道类型是(*io.Reader)(nil)
		// reflect.Type: *io.Reader
		// reflect.Value: nil
		// 导致即使Apply根据type,name找到value, 但是由于value的IsValid返回的false导致注入失败
		// 所以此处改为判断根据type,name能否找到value，而不关注是否是IsValid
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

			if val := inj.Get(tt, injectName); !val.IsValid() {
				errs = append(errs, MissingDependencyError{
					Field:       structField.Name,
					ReflectType: tt.String(),
					InjectName:  injectName,
				})
			}
		}

	}
	return errs
}

func checkOut(inj inject.Injector, t reflect.Type) []error {
	var errs []error
	for i := 0; i < t.NumOut(); i++ {
		outType := t.Out(i)

		if outType.Implements(errorInterface) {
			continue
		}

		for outType.Kind() == reflect.Ptr {
			outType = outType.Elem()
		}

		if outType.Kind() != reflect.Struct {
			errs = append(errs, fmt.Errorf("Cannot support types other than structures %v", outType))
		}

		val := reflect.New(outType)
		// 接口类型 (*io.Reader)(nil)
		// 基础类型 (string)("")
		// 结构体指针类型 (*Foo)(nil)
		// 结构体类型 (Foo)({})
		// 由于check流程是直接反射方法造处对应接口，无法或者接口类型的具体value
		if err := inj.MapValues(val); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func filterMissingDependencyError(errs []error) []MissingDependencyError {
	var mdeErrs []MissingDependencyError
	for _, err := range errs {
		cause, ok := errors.Cause(err).(MissingDependencyError)
		if ok {
			mdeErrs = append(mdeErrs, cause)
		}
	}
	return mdeErrs
}
