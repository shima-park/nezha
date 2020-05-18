package component

import "reflect"

type Component interface {
	// 组件实例配置
	SampleConfig() string

	// 组件描述
	Description() string

	// 获取组件实例对象
	Instance() Instance

	// 组件启动
	Start() error

	// 组件停止
	Stop() error
}

type Instance interface {
	Name() string
	// 组件的Go Type
	Type() reflect.Type
	// 组件的Go Type
	Value() reflect.Value
	// 组件的实例
	Interface() interface{}
}

type instance struct {
	name  string
	typ   reflect.Type
	value reflect.Value
	iface interface{}
}

func NewInstance(name string, typ reflect.Type, value reflect.Value, iface interface{}) Instance {
	return &instance{
		name:  name,
		typ:   typ,
		value: value,
		iface: iface,
	}
}

func (i *instance) Name() string {
	return i.name
}

func (i *instance) Type() reflect.Type {
	return i.typ
}

func (i *instance) Value() reflect.Value {
	return i.value
}

func (i *instance) Interface() interface{} {
	return i.iface
}
