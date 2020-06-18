package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var cmdComponent = &cobra.Command{
	Use:     "component",
	Aliases: []string{"comp"},
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var cmdCompList = &cobra.Command{
	Use:   "list",
	Short: "Display one or many component",
	Run: func(cmd *cobra.Command, args []string) {
		list, err := newClient().Component.List()
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
	},
}

var cmdCompConfig = &cobra.Command{
	Use:     "config",
	Aliases: []string{"conf"},
	Short:   "Display config of component",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("missing processor name")
			os.Exit(1)
		}
		c := newClient()
		for _, name := range args {
			conf, err := c.Component.Config(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(conf)
		}
	},
}

func init() {
	cmdComponent.AddCommand(cmdCompConfig, cmdCompList)
	rootCmd.AddCommand(cmdComponent)
}
