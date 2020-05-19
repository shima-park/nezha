package pipeline

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/shima-park/inject"
	"github.com/shima-park/nezha/pkg/processor"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

type execContext struct {
	ctx      context.Context
	injector inject.Injector
	stream   *Stream
}

func NewExecContext(ctx context.Context, s *Stream, parent inject.Injector) *execContext {
	inj := inject.New()
	inj.SetParent(parent)
	inj.MapTo(ctx, "Context", (*context.Context)(nil))

	return &execContext{
		ctx:      ctx,
		injector: inj,
		stream:   s,
	}
}

func (c *execContext) Run() error {
	return run(c.stream, c.injector)
}

func run(s *Stream, injector inject.Injector) error {
	if err := invoke(s, injector); err != nil {
		return err
	}

	for i := 0; i < len(s.childs); i++ {
		if err := run(s.childs[i], injector); err != nil {
			return err
		}
	}

	return nil
}

func invoke(s *Stream, injector inject.Injector) error {
	if s == nil || s.processor == nil {
		return nil
	}

	if err := processor.Validate(s.processor); err != nil {
		return err
	}

	vals, err := injector.Invoke(s.processor)
	if err != nil {
		return errors.Wrapf(err, "Stream(%s)", s.Name())
	}

	if err = setInjector(injector, vals...); err != nil {
		return errors.Wrapf(err, "Stream(%s)", s.Name())
	}
	return nil
}

func setInjector(inj inject.Injector, vals ...reflect.Value) error {
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

		for i := 0; i < val.NumField(); i++ {
			f := val.Field(i)
			structField := typ.Field(i)
			injectName := getInjectName(structField)

			if f.Type().Kind() == reflect.Interface {
				nilPtr := reflect.New(f.Type())
				inj.MapTo(f.Interface(), injectName, nilPtr.Interface())
			} else {
				inj.Set(f.Type(), injectName, f)
			}
		}
	}
	return nil
}

func getInjectName(structField reflect.StructField) string {
	var injectName = structField.Name
	if structField.Tag.Get("inject") != "" {
		injectName = structField.Tag.Get("inject")
	}
	return injectName
}

func check(s *Stream, inj inject.Injector) error {
	if s == nil || s.processor == nil {
		return nil
	}

	if err := processor.Validate(s.processor); err != nil {
		return errors.Wrapf(err, "Stream(%s)", s.Name())
	}

	if err := checkDep(inj, s.processor); err != nil {
		return errors.Wrapf(err, "Stream(%s)", s.Name())
	}

	for i := 0; i < len(s.childs); i++ {
		if err := check(s.childs[i], inj); err != nil {
			return err
		}
	}
	return nil
}

func checkDep(inj inject.Injector, f interface{}) error {
	t := reflect.TypeOf(f)

	if err := checkIn(inj, t); err != nil {
		return err
	}

	if err := checkOut(inj, t); err != nil {
		return err
	}

	return nil
}

func checkIn(inj inject.Injector, t reflect.Type) error {
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)

		for argType.Kind() == reflect.Ptr {
			argType = argType.Elem()
		}

		if argType.Kind() != reflect.Struct {
			return fmt.Errorf("Cannot support types other than structures %v", argType)
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
			injectName := getInjectName(structField)

			var tt reflect.Type
			if f.Type().Kind() == reflect.Interface {
				nilPtr := reflect.New(f.Type())
				tt = inject.InterfaceOf(nilPtr.Interface())
			} else {
				tt = f.Type()
			}

			if ok := inj.Exists(tt, injectName); !ok {
				return fmt.Errorf("Value not found for type: %v name: %v", tt, injectName)
			}
		}

	}
	return nil
}

func checkOut(inj inject.Injector, t reflect.Type) error {
	for i := 0; i < t.NumOut(); i++ {
		outType := t.Out(i)

		if outType.Implements(errorInterface) {
			continue
		}

		for outType.Kind() == reflect.Ptr {
			outType = outType.Elem()
		}

		if outType.Kind() != reflect.Struct {
			return fmt.Errorf("Cannot support types other than structures %v", outType)
		}

		val := reflect.New(outType)
		// 接口类型 (*io.Reader)(nil)
		// 基础类型 (string)("")
		// 结构体指针类型 (*Foo)(nil)
		// 结构体类型 (Foo)({})
		// 由于check流程是直接反射方法造处对应接口，无法或者接口类型的具体value
		if err := setInjector(inj, val); err != nil {
			return err
		}
	}
	return nil
}
