package io

import (
	"io"
	"nezha/pkg/component"
	"nezha/pkg/common/config"
	"os"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/shima-park/inject"
)

var _ component.Component = &Reader{}

func init() {
	if err := component.Register("io_reader", func(config string) (component.Component, error) {
		return NewReader(config)
	}); err != nil {
		panic(err)
	}
}

type ReaderConfig struct {
	Name string
	Path string
}

type Reader struct {
	rc     io.ReadCloser
	config ReaderConfig
}

func NewReader(rawConfig string) (*Reader, error) {
	var conf ReaderConfig
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
		rc:     f,
		config: conf,
	}, nil
}

func (r *Reader) SampleConfig() string {
	conf := &ReaderConfig{
		Path: "stdin",
	}

	b, _ := config.Marshal(conf)

	return string(b)
}

func (r *Reader) Description() string {
	return "file reader e.g.: stdin, stdout, stderr, /var/log/xxx.log"
}

func (r *Reader) Instance() component.Instance {
	return component.Instance{
		Name:      r.config.Name,
		Type:      inject.InterfaceOf((*io.Reader)(nil)),
		Value:     reflect.ValueOf(r.rc),
		Interface: r.rc,
	}
}

func (r *Reader) Start() error {
	return nil
}

func (r *Reader) Stop() error {
	return r.rc.Close()
}
