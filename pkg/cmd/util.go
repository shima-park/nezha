package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/shima-park/nezha/pkg/rpc/client"
)

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

func handleErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func stringInSlice(t string, ss []string) bool {
	for _, s := range ss {
		if t == s {
			return true
		}
	}
	return false
}
