package io

import (
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/shima-park/nezha/common/config"
	"github.com/shima-park/nezha/component"

	"github.com/pkg/errors"
	"github.com/shima-park/nezha/common/inject"
)

var (
	readerFactory       component.Factory   = NewReaderFactory()
	_                   component.Component = &Reader{}
	defaultReaderConfig                     = ReaderConfig{
		Name: "MyReader",
		Path: "stdin",
	}
	readerDescription = "file reader e.g.: stdin, stdout, stderr, /var/log/xxx.log"
)

func init() {
	if err := component.Register("io_reader", readerFactory); err != nil {
		panic(err)
	}
}

func NewReaderFactory() component.Factory {
	return component.NewFactory(
		defaultReaderConfig,
		readerDescription,
		func(c string) (component.Component, error) {
			return NewReader(c)
		})
}

type ReaderConfig struct {
	Name string
	Path string
}

type Reader struct {
	rc       io.ReadCloser
	instance component.Instance
}

func NewReader(rawConfig string) (*Reader, error) {
	conf := defaultReaderConfig
	err := config.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	var f io.ReadCloser
	switch strings.TrimSpace(conf.Path) {
	case "stdin":
		f = os.Stdin
	case "stdout":
		f = os.Stdout
	case "stderr":
		f = os.Stderr
	default:
		file, err := os.Open(conf.Path)
		if err != nil {
			return nil, errors.Wrap(err, "io_reader")
		}
		f = file
	}

	return &Reader{
		rc: f,
		instance: component.NewInstance(
			conf.Name,
			inject.InterfaceOf((*io.Reader)(nil)),
			reflect.ValueOf(f),
			f,
		),
	}, nil
}

func (r *Reader) Instance() component.Instance {
	return r.instance
}

func (r *Reader) Start() error {
	return nil
}

func (r *Reader) Stop() error {
	return r.rc.Close()
}
