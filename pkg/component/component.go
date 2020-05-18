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

type Instance struct {
	// 组件的实例名字
	Name string
	// 组件的Go Type
	Type reflect.Type
	// 组件的Go Type
	Value reflect.Value
	// 组件的实例
	Interface interface{}
}
