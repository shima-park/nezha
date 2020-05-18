package main

import (
	"context"
	"encoding/json"
	"io"

	"nezha/pkg/common/plugin"
	"nezha/pkg/processor"
	"nezha/test/proto"
)

var Bundle = plugin.Bundle(
	processor.Plugin("jsondecode_str_2_foo", ProcessorFactory),
)

type Request struct {
	Ctx    context.Context `inject`
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
