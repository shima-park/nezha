package pipeline

import (
	"context"
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
