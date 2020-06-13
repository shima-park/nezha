package inject

import (
	"reflect"

	"github.com/shima-park/nezha/common/log"
)

type loggerInjector struct {
	where string
	Injector
}

func NewLoggerInjector(where string, injector Injector) Injector {
	return &loggerInjector{
		where:    where,
		Injector: injector,
	}
}

func (i *loggerInjector) Map(value interface{}, name string) TypeMapper {
	log.Info("%s Map type: %s, name: %s to injector",
		i.where, reflect.TypeOf(value), name)
	i.Injector.Map(value, name)
	return i
}

func (i *loggerInjector) MapTo(value interface{}, name string, ifacePtr interface{}) TypeMapper {
	log.Info("%s MapTo type: %s, name: %s to injector",
		i.where, reflect.TypeOf(ifacePtr), name)
	i.Injector.MapTo(value, name, ifacePtr)
	return i
}

func (i *loggerInjector) Set(typ reflect.Type, name string, value reflect.Value) TypeMapper {
	log.Info("%s Set type: %s, name: %s to injector",
		i.where, typ, name)
	i.Injector.Set(typ, name, value)
	return i
}
