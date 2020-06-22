package cmd

import (
	"fmt"

	"github.com/shima-park/nezha/pkg/rpc/proto"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewGetCmd(cmds ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get (RESOURCE/NAME | -p PIPELINE_NAME)",
		Short: "Display one or many resources",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(cmds...)
	return cmd
}

func NewGetPipeCmd() *cobra.Command {
	var o string
	cmd := &cobra.Command{
		Use:     "pipeline",
		Aliases: []string{"pipe"},
		Short:   "Display pipeline list",
		Run: func(cmd *cobra.Command, args []string) {
			list, err := newClient().Pipeline.List()
			handleErr(err)

			var filters []proto.PipelineView
			for _, e := range list {
				if len(args) > 0 && !stringInSlice(e.Name, args) {
					continue
				}
				filters = append(filters, e)
			}

			if o == "" {
				var rows [][]string
				for _, e := range filters {
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
			} else {
				for _, e := range filters {
					fmt.Println(string(e.RawConfig))
				}
			}
		},
	}

	cmd.Flags().StringVarP(&o, "output", "o", "", "Output by yaml format.")

	return cmd
}

func NewGetCompCmd() *cobra.Command {
	var p string
	cmd := &cobra.Command{
		Use:     "component",
		Aliases: []string{"comp"},
		Short:   "Display component list",
		Run: func(cmd *cobra.Command, args []string) {
			var list []proto.ComponentView
			if p == "" {
				fmt.Println("Register components:(show pipeline's components use -p {pipeline_name})")
				var err error
				list, err = newClient().Component.List()
				handleErr(err)
			} else {
				fmt.Printf("Pipeline: %s components:\n", p)
				pipe, err := newClient().Pipeline.Find(p)
				handleErr(err)
				list = pipe.Components
			}

			var rows [][]string
			for _, e := range list {
				if len(args) > 0 && !stringInSlice(e.Name, args) {
					continue
				}
				rows = append(rows, []string{
					e.Name, e.RawConfig, e.SampleConfig, e.Description, e.InjectName, e.ReflectType,
				})
			}

			header := []string{
				"name", "raw_config", "sample_config", "desc", "inject_name", "reflect_type",
			}

			renderTable(header, rows)
		},
	}

	cmd.Flags().StringVarP(&p, "pipeline", "p", "", "The pipeline scope for this CLI request")

	return cmd
}

func NewGetProcCmd() *cobra.Command {
	var p string
	cmd := &cobra.Command{
		Use:     "processor",
		Aliases: []string{"proc"},
		Short:   "Display processor list",
		Run: func(cmd *cobra.Command, args []string) {
			var list []proto.ProcessorView
			if p == "" {
				fmt.Println("Register processors:(show pipeline's processors use -p {pipeline_name})")
				var err error
				list, err = newClient().Processor.List()
				handleErr(err)
			} else {
				fmt.Printf("Pipeline: %s processors:\n", p)
				pipe, err := newClient().Pipeline.Find(p)
				handleErr(err)
				list = pipe.Processors
			}

			var rows [][]string
			for _, e := range list {
				if len(args) > 0 && !stringInSlice(e.Name, args) {
					continue
				}
				rows = append(rows, []string{
					e.Name, e.RawConfig, e.SampleConfig, e.Description,
				})
			}

			header := []string{
				"name", "raw_config", "sample_config", "desc",
			}

			renderTable(header, rows)
		},
	}

	cmd.Flags().StringVarP(&p, "pipeline", "p", "", "The pipeline scope for this CLI request")

	return cmd
}

func NewGetPluginCmd() *cobra.Command {
	var p string
	cmd := &cobra.Command{
		Use:     "plugin",
		Aliases: []string{"plug"},
		Short:   "Display plugin list",
		Run: func(cmd *cobra.Command, args []string) {
			list, err := newClient().Plugin.List()
			handleErr(err)

			var rows [][]string
			for _, e := range list {
				if len(args) > 0 && !stringInSlice(e.Path, args) {
					continue
				}
				rows = append(rows, []string{
					e.Path, e.Module, e.OpenTime,
				})
			}

			header := []string{"path", "module", "open_time"}

			renderTable(header, rows)
		},
	}

	cmd.Flags().StringVar(&p, "p", "", "The pipeline scope for this CLI request")

	return cmd
}

func NewGetServerCmd() *cobra.Command {
	var p string
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"serv"},
		Short:   "Display server metadata",
		Run: func(cmd *cobra.Command, args []string) {
			meta, err := newClient().Server.Metadata()
			handleErr(err)

			b, err := yaml.Marshal(meta)
			handleErr(err)
			fmt.Println(string(b))
		},
	}

	cmd.Flags().StringVar(&p, "p", "", "The pipeline scope for this CLI request")

	return cmd
}

func init() {
	rootCmd.AddCommand(
		NewGetCmd(
			NewGetPipeCmd(), NewGetCompCmd(), NewGetProcCmd(), NewGetPluginCmd(),
			NewGetServerCmd(),
		),
	)
}
