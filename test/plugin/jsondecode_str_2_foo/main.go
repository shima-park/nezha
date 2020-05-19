package main

import (
	"context"
	"encoding/json"
	"io"

	"github.com/shima-park/nezha/pkg/common/plugin"
	"github.com/shima-park/nezha/pkg/processor"
	"github.com/shima-park/nezha/test/proto"
)

var Bundle = plugin.Bundle(
	processor.Plugin("jsondecode_str_2_foo", ProcessorFactory),
)

type Request struct {
	Ctx    context.Context `inject:"Context"`
	Msg    string          `inject:"Message"`
	Reader io.Reader       `inject:"Reader"`
}

type Response struct {
	Foo proto.Foo `inject`
}

func Handle(r Request) (Response, error) {
	var f proto.Foo
	err := json.Unmarshal([]byte(r.Msg), &f)
	return Response{Foo: f}, err
}

func ProcessorFactory(config string) (processor.Processor, error) {
	return Handle, nil
}
