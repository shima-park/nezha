package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/shima-park/nezha/pkg/common/plugin"
	_ "github.com/shima-park/nezha/pkg/component/include"
	"github.com/shima-park/nezha/pkg/pipeline"
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
		Pipeline: pipeline.PipelineConfig{
			Stream: &pipeline.StreamConfig{
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
		},
	}

	c, err := pipeline.NewPipelineByConfig(pipeConf)
	if err != nil {
		panic(err)
	}

	fmt.Println("Components:")
	pipeline.PrintPipelineComponents(os.Stdout, c)
	fmt.Println("Processor:")
	pipeline.PrintPipelineProcessor(os.Stdout, c)

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
