package io

import (
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/shima-park/nezha/pkg/common/config"
	"github.com/shima-park/nezha/pkg/component"

	"github.com/pkg/errors"
	"github.com/shima-park/nezha/pkg/inject"
)

var (
	writerFactory       component.Factory   = NewWriterFactory()
	_                   component.Component = &Writer{}
	defaultWriterConfig                     = ReaderConfig{
		Name: "MyWriter",
		Path: "stdout",
	}
	writerDescription = "file writer e.g.: stdout, stderr, /dev/null, /var/log/xxx.log"
)

func init() {
	if err := component.Register("io_writer", writerFactory); err != nil {
		panic(err)
	}
}

func NewWriterFactory() component.Factory {
	return component.NewFactory(
		defaultWriterConfig,
		writerDescription,
		func(c string) (component.Component, error) {
			return NewWriter(c)
		})
}

type WriterConfig struct {
	Name string
	Path string
}

type Writer struct {
	wc       io.WriteCloser
	instance component.Instance
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func NopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}

func NewWriter(rawConfig string) (*Writer, error) {
	conf := defaultWriterConfig
	err := config.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	var f io.WriteCloser
	switch strings.TrimSpace(conf.Path) {
	case "/dev/null":
		f = NopCloser(ioutil.Discard)
	case "stdout":
		f = os.Stdout
	case "stderr":
		f = os.Stderr
	default:
		file, err := os.Open(conf.Path)
		if err != nil {
			return nil, errors.Wrap(err, "io_writer")
		}
		f = file
	}

	return &Writer{
		wc: f,
		instance: component.NewInstance(
			conf.Name,
			inject.InterfaceOf((*io.Writer)(nil)),
			reflect.ValueOf(f),
			f,
		),
	}, nil
}

func (w *Writer) Instance() component.Instance {
	return w.instance
}

func (w *Writer) Start() error {
	return nil
}

func (w *Writer) Stop() error {
	return w.wc.Close()
}
