package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var cmdProcessor = &cobra.Command{
	Use:     "processor",
	Aliases: []string{"proc"},
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var cmdProcList = &cobra.Command{
	Use:   "list",
	Short: "Display one or many processor",
	Run: func(cmd *cobra.Command, args []string) {
		list, err := newClient().Processor.List()
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
	},
}

var cmdProcConfig = &cobra.Command{
	Use:     "config",
	Aliases: []string{"conf"},
	Short:   "Display config of processor",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("missing processor name")
			os.Exit(1)
		}
		c := newClient()
		for _, name := range args {
			conf, err := c.Processor.Config(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(conf)
		}
	},
}

func init() {
	cmdProcessor.AddCommand(cmdProcConfig, cmdProcList)
	rootCmd.AddCommand(cmdProcessor)
}
