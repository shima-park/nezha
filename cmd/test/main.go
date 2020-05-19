package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/shima-park/nezha/pkg/common/log"
	"github.com/shima-park/nezha/pkg/common/plugin"
	_ "github.com/shima-park/nezha/pkg/component/include"
	"github.com/shima-park/nezha/pkg/pipeline"
)

var plugins = &pluginList{}

func init() {
	flag.Var(plugins, "plugin", "Load additional plugins")
}

func Initialize() error {
	for _, path := range plugins.paths {
		log.Info("loading plugin bundle: %v", path)

		if err := plugin.LoadPlugins(path); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	flag.Parse()

	if err := Initialize(); err != nil {
		panic(err)
	}

	pipeConf := pipeline.Config{
		Name: "test_pipeline",
		Components: []map[string]string{
			{"io_reader": `{"Name":"Stdin","Path":"stdin"}`},
			{"io_writer": `{"Name":"Stdout","Path":"stdout"}`},
		},
	}

	pipeConf.
		AddStream(pipeline.StreamConfig{
			ProcessorName: "read_line_from_stdin",
		}).
		AddStream(pipeline.StreamConfig{
			ProcessorName: "jsondecode_str_2_foo",
		}).
		AddStream(pipeline.StreamConfig{
			ProcessorName: "write_foo_2_stdout",
		})

	c, err := pipeline.NewPipelineByConfig(pipeConf)
	if err != nil {
		panic(err)
	}

	fmt.Println("Components:\n", strings.Join(c.ListComponent(), "\n"))

	if err := c.Start(); err != nil {
		panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	<-signals

	c.Stop()
}

type pluginList struct {
	paths []string
}

func (p *pluginList) String() string {
	return strings.Join(p.paths, ",")
}

func (p *pluginList) Set(v string) error {
	for _, path := range p.paths {
		if path == v {
			log.Warn("%s is already a registered plugin", path)
			return nil
		}
	}
	p.paths = append(p.paths, v)
	return nil
}
