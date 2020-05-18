package pipeline

import (
	"context"
	"github.com/shima-park/nezha/pkg/common/log"
	"github.com/shima-park/nezha/pkg/processor"
	"reflect"

	"github.com/shima-park/inject"
)

type execContext struct {
	ctx      context.Context
	injector inject.Injector
	stream   *Stream
}

func NewExecContext(ctx context.Context, s *Stream, parent inject.Injector) *execContext {
	inj := inject.New()
	inj.SetParent(parent)
	inj.MapTo(ctx, "Ctx", (*context.Context)(nil))

	return &execContext{
		ctx:      ctx,
		injector: inj,
		stream:   s,
	}
}

func (c *execContext) Run() error {
	return run(c.ctx, c.stream, c.injector)
}

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

func run(ctx context.Context, s *Stream, injector inject.Injector) error {
	if s == nil || s.processor == nil {
		return nil
	}

	if err := processor.Validate(s.processor); err != nil {
		return err
	}

	vals, err := injector.Invoke(s.processor)
	if err != nil {
		return err
	}

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
			var injectName = structField.Name
			if structField.Tag.Get("inject") != "" {
				injectName = structField.Tag.Get("inject")
			}

			if f.Type().Kind() == reflect.Interface {
				nilPtr := reflect.New(f.Type())
				log.Info("Add Type: %s, Name: %s, Value: %v to injector\n",
					f.Type(), injectName, nilPtr.Interface())
				injector.MapTo(f.Interface(), injectName, nilPtr.Interface())
			} else {
				log.Info("Add Type: %s, Name: %s, Value: %v to injector\n",
					f.Type(), injectName, f)
				injector.Set(f.Type(), injectName, f)
			}
		}
	}

	for i := 0; i < len(s.childs); i++ {
		sf := s.childs[i]
		if err = run(ctx, sf, injector); err != nil {
			return err
		}
	}
	return nil
}
