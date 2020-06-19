package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/shima-park/lotus/pipeline"
	"github.com/shima-park/nezha/rpc/proto"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var cmdPipeline = &cobra.Command{
	Use:     "pipeline",
	Aliases: []string{"pipe"},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var cmdPipeList = &cobra.Command{
	Use:   "list",
	Short: "Display one or many pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		list, err := newClient().Pipeline.List()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var rows [][]string
		for _, e := range list {
			rows = append(rows, []string{e.Name, e.State, e.Schedule, fmt.Sprint(e.Bootstrap),
				e.StartTime, e.ExitTime, e.RunTimes, e.NextRunTime, e.LastStartTime, e.LastEndTime})
		}

		renderTable(
			[]string{
				"name", "state", "schedule", "bootstrap", "start_time", "exit_time",
				"run_times", "next_run_time", "last_start_time", "last_end_time",
			},
			rows,
		)
	},
}

var cmdPipeConfig = &cobra.Command{
	Use:     "config",
	Aliases: []string{"conf"},
	Short:   "Display config of pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("missing processor name")
			os.Exit(1)
		}
		c := newClient()
		for _, name := range args {
			conf, err := c.Pipeline.Config(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(conf)
		}
	},
}

var cmdStartPipe = &cobra.Command{
	Use:   "start",
	Short: "start a pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		err := newClient().Pipeline.Control(proto.ControlCommandStart, args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var cmdStopPipe = &cobra.Command{
	Use:   "stop",
	Short: "stop a pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		err := newClient().Pipeline.Control(proto.ControlCommandStop, args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var cmdRestartPipe = &cobra.Command{
	Use:   "restart",
	Short: "restart a pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		err := newClient().Pipeline.Control(proto.ControlCommandRestart, args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var cmdPipeVars = &cobra.Command{
	Use:   "vars",
	Short: "show vars of pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		c := newClient()
		for _, name := range args {
			vars, err := c.Pipeline.Vars(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{
				"key", "value",
			})
			table.SetRowLine(true)
			for k, v := range vars {
				table.Append([]string{
					k, v,
				})
			}
			table.Render()
		}
	},
}

var cmdPipeComponents = &cobra.Command{
	Use:     "components",
	Aliases: []string{"comp"},
	Short:   "show components of pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		c := newClient()
		for _, name := range args {
			list, err := c.Pipeline.ListComponents(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{
				"name", "raw_config", "sample_config", "desc", "inject_name", "reflect_type",
			})
			table.SetRowLine(true)
			for _, e := range list {
				table.Append([]string{
					e.Name, e.RawConfig, e.SampleConfig, e.Description, e.InjectName, e.ReflectType,
				})
			}
			table.Render()
		}
	},
}

var cmdPipeProcessors = &cobra.Command{
	Use:     "processors",
	Aliases: []string{"proc"},
	Short:   "show processors of pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		c := newClient()
		for _, name := range args {
			list, err := c.Pipeline.ListProcessors(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{
				"name", "raw_config", "sample_config", "desc",
			})
			table.SetRowLine(true)
			for _, e := range list {
				table.Append([]string{
					e.Name,
					e.RawConfig,
					e.SampleConfig,
					e.Description,
				})
			}
			table.Render()
		}
	},
}

func init() {
	var name string
	var processors []string
	var components []string
	var cmdPipeGenConf = &cobra.Command{
		Use:   "gf",
		Short: "generate config of pipeline",
		Run: func(cmd *cobra.Command, args []string) {
			c := newClient()
			conf, err := c.Pipeline.GenerateConfig(name, components, processors)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(conf)
		},
	}
	cmdPipeGenConf.Flags().StringVar(&name, "name", "", "name of pipeline")
	cmdPipeGenConf.Flags().StringSliceVar(&processors, "processors", nil, "processors of pipeline")
	cmdPipeGenConf.Flags().StringSliceVar(&components, "components", nil, "components of pipeline")

	var configPath string
	var rawConfig string
	var cmdAddPipe = &cobra.Command{
		Use:   "add",
		Short: "add a pipeline",
		Run: func(cmd *cobra.Command, args []string) {

			var data []byte
			if configPath != "" {
				var err error
				data, err = ioutil.ReadFile(configPath)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			} else if rawConfig != "" {
				data = []byte(rawConfig)
			} else {
				fmt.Println("--path or --raw you at least provide one of them")
			}

			var conf pipeline.Config
			err := yaml.Unmarshal(data, &conf)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = newClient().Pipeline.Add(conf)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	cmdAddPipe.Flags().StringVar(&configPath, "path", "", "path to pipeline config")
	cmdAddPipe.Flags().StringVar(&rawConfig, "raw", "", "raw of pipeline config")

	cmdPipeline.AddCommand(
		cmdStartPipe, cmdStopPipe, cmdRestartPipe,
		cmdAddPipe, cmdPipeVars, cmdPipeGenConf,
		cmdPipeProcessors, cmdPipeComponents, cmdPipeConfig,
		cmdPipeList,
	)
	rootCmd.AddCommand(cmdPipeline)
}
