package io

import (
	"io"
	"io/ioutil"
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
	if err := component.Register("io_writer", func(config string) (component.Component, error) {
		return NewWriter(config)
	}); err != nil {
		panic(err)
	}
}

type WriterConfig struct {
	Name string
	Path string
}

type Writer struct {
	wc     io.WriteCloser
	config WriterConfig
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func NopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}

func NewWriter(rawConfig string) (*Writer, error) {
	var conf WriterConfig
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
		wc:     f,
		config: conf,
	}, nil
}

func (w *Writer) SampleConfig() string {
	conf := &ReaderConfig{
		Path: "stdout",
	}

	b, _ := config.Marshal(conf)

	return string(b)
}

func (w *Writer) Description() string {
	return "file writer e.g.: stdout, stderr, /dev/null, /var/log/xxx.log"
}

func (w *Writer) Instance() component.Instance {
	return component.Instance{
		Name:      w.config.Name,
		Type:      inject.InterfaceOf((*io.Writer)(nil)),
		Value:     reflect.ValueOf(w.wc),
		Interface: w.wc,
	}
}

func (w *Writer) Start() error {
	return nil
}

func (w *Writer) Stop() error {
	return w.wc.Close()
}
