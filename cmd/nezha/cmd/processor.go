package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cmdProcessor = &cobra.Command{
	Use:     "processor",
	Aliases: []string{"proc"},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
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

		var rows [][]string
		for _, e := range list {
			rows = append(rows, []string{
				e.Name, e.RawConfig, e.SampleConfig, e.Description,
			})
		}

		renderTable(
			[]string{"name", "raw_config", "sample_config", "desc"},
			rows,
		)
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
