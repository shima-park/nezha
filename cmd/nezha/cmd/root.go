package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/shima-park/nezha/rpc/client"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nezha",
	Short: "nezha is a pipeline-based task scheduling center",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newClient() *client.Client {
	return client.NewClient("localhost:8080")
}

func renderTable(header []string, rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}
