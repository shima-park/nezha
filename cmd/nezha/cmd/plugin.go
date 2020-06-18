package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var cmdPlugin = &cobra.Command{
	Use:     "plugin",
	Aliases: []string{"plug"},
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var cmdPluginList = &cobra.Command{
	Use:   "list",
	Short: "Display one or many plugin",
	Run: func(cmd *cobra.Command, args []string) {
		list, err := newClient().Plugin.List()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{
			"path", "module", "open_time",
		})
		table.SetRowLine(true)
		for _, e := range list {
			table.Append([]string{
				e.Path, e.Module, e.OpenTime,
			})
		}
		table.Render()
	},
}

var cmdPluginOpen = &cobra.Command{
	Use:   "open",
	Short: "open a new plugin",
	Run: func(cmd *cobra.Command, args []string) {
		for _, path := range args {
			err := newClient().Plugin.Open(path)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	},
}

var cmdPluginAdd = &cobra.Command{
	Use:   "add",
	Short: "add a new plugin",
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
			err := c.Plugin.Upload(path)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	cmdPlugin.AddCommand(cmdPluginList, cmdPluginOpen, cmdPluginAdd)
	rootCmd.AddCommand(cmdPlugin)
}
