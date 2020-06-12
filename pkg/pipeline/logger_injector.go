package pipeline

import (
	"reflect"

	"github.com/shima-park/nezha/pkg/common/log"
	"github.com/shima-park/nezha/pkg/inject"
)

type loggerInjector struct {
	where string
	inject.Injector
}

func NewLoggerInjector(where string, injector inject.Injector) inject.Injector {
	return &loggerInjector{
		where:    where,
		Injector: injector,
	}
}

func (i *loggerInjector) Map(value interface{}, name string) inject.TypeMapper {
	log.Info("%s Map type: %s, name: %s to injector",
		i.where, reflect.TypeOf(value), name)
	i.Injector.Map(value, name)
	return i
}

func (i *loggerInjector) MapTo(value interface{}, name string, ifacePtr interface{}) inject.TypeMapper {
	log.Info("%s MapTo type: %s, name: %s to injector",
		i.where, reflect.TypeOf(ifacePtr), name)
	i.Injector.MapTo(value, name, ifacePtr)
	return i
}

func (i *loggerInjector) Set(typ reflect.Type, name string, value reflect.Value) inject.TypeMapper {
	log.Info("%s Set type: %s, name: %s to injector",
		i.where, typ, name)
	i.Injector.Set(typ, name, value)
	return i
}
