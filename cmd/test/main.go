package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/shima-park/nezha/common/plugin"
	_ "github.com/shima-park/nezha/component/include"
	"github.com/shima-park/nezha/pipeline"
)

func main() {
	flag.Parse()

	if err := plugin.Initialize(); err != nil {
		panic(err)
	}

	pipeConf := pipeline.Config{
		Name: "test_pipeline",
		Components: []map[string]string{
			{"io_reader": `
name: Stdin
path: stdin`},
			{"io_writer": `
name: Stdout
path: stdout`},
		},
		Stream: pipeline.StreamConfig{
			Name: "read_line_from_stdin",
			Childs: []pipeline.StreamConfig{
				pipeline.StreamConfig{
					Name: "jsondecode_str_2_foo",
					Childs: []pipeline.StreamConfig{
						pipeline.StreamConfig{
							Name: "write_foo_2_stdout",
						},
					},
				},
			},
		},
	}

	c, err := pipeline.NewPipelineByConfig(pipeConf)
	if err != nil {
		panic(err)
	}

	err = c.Visualize(os.Stdout, "ascii_table")
	if err != nil {
		panic(err)
	}

	if errs := c.CheckDependence(); len(errs) > 0 {
		panic(errs[0])
	}

	if err := c.Start(); err != nil {
		panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	<-signals

	c.Stop()
}
