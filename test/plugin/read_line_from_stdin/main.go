package main

import (
	"bufio"
	"io"

	"nezha/pkg/common/plugin"
	"nezha/pkg/processor"
)

var Bundle = plugin.Bundle(
	processor.Plugin("read_line_from_stdin", ProcessorFactory),
)

type Request struct {
	Reader io.Reader `inject:"Stdin"`
}

type Response struct {
	Msg    string    `inject:"Message"`
	Reader io.Reader `inject:"Reader"`
}

func Handle(r Request) (Response, error) {
	str, err := bufio.NewReader(r.Reader).ReadString('\n')
	return Response{
		Msg:    str,
		Reader: r.Reader,
	}, err
}

func ProcessorFactory(config string) (processor.Processor, error) {
	return Handle, nil
}
