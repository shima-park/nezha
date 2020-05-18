package main

import (
	"fmt"
	"io"

	"github.com/shima-park/nezha/pkg/common/plugin"
	"github.com/shima-park/nezha/pkg/processor"
	"github.com/shima-park/nezha/test/proto"
)

var Bundle = plugin.Bundle(
	processor.Plugin("write_foo_2_stdout", ProcessorFactory),
)

type Request struct {
	Writer io.Writer `inject:"Stdout"`
	Foo    proto.Foo `inject:"Foo"`
}

type Response struct {
}

func Handle(r Request) error {
	fmt.Fprintf(r.Writer, "Your Foo: Name: %s, Age: %d\n", r.Foo.Name, r.Foo.Age)
	return nil
}

func ProcessorFactory(config string) (processor.Processor, error) {
	return Handle, nil
}
