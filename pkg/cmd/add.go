package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/shima-park/lotus/pipeline"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewAddCmd(cmds ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add (RESOURCE/NAME | -f FILENAME)",
		Short: "add a resource to the server",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(cmds...)
	return cmd
}

func NewAddPipelineCmd() *cobra.Command {
	var file string
	var name string
	var schedule string
	var processors []string
	var components []string
	cmd := &cobra.Command{
		Use:     "pipeline (NAME)",
		Aliases: []string{"pipe"},
		Short:   "Add a pipeline to the server",
		Run: func(cmd *cobra.Command, args []string) {
			if file != "" {
				data, err := ioutil.ReadFile(file)
				handleErr(err)

				var conf pipeline.Config
				err = yaml.Unmarshal(data, &conf)
				handleErr(err)

				err = newClient().Pipeline.Add(conf)
				handleErr(err)
			} else if name != "" || len(processors) > 0 || len(components) > 0 {
				c := newClient()
				conf, err := c.Pipeline.GenerateConfig(name, schedule, components, processors)
				handleErr(err)

				origin, err := yaml.Marshal(conf)
				handleErr(err)

				err = runEditor(origin, c.Pipeline.Add)
				if err != nil {
					handleErr(err)
				}
			} else {
				fmt.Println("-f pipeline.yaml or -n test -p test_print -c es_client you at least provide one of them")
			}

		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to pipeline config")
	cmd.Flags().StringVarP(&name, "name", "n", "", "name of pipeline")
	cmd.Flags().StringVarP(&schedule, "schedule", "s", "", "name of pipeline")
	cmd.Flags().StringSliceVarP(&processors, "processors", "p", nil, "processors of pipeline")
	cmd.Flags().StringSliceVarP(&components, "components", "c", nil, "components of pipeline")

	return cmd
}

func NewAddPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugin (PATH)",
		Aliases: []string{"plug"},
		Short:   "Add a plugin to the server",
		Run: func(cmd *cobra.Command, args []string) {
			for _, path := range args {
				_, err := os.Lstat(path)
				if os.IsNotExist(err) {
					fmt.Println(err)
					os.Exit(1)
				}
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
			c := newClient()
			for _, path := range args {
				err := c.Plugin.Add(path)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		},
	}
	return cmd
}

func init() {
	rootCmd.AddCommand(
		NewAddCmd(
			NewAddPipelineCmd(), NewAddPluginCmd(),
		),
	)
}
