package cmd

import (
	"os"
	"os/signal"

	"github.com/shima-park/nezha/pkg/rpc/server"
	"github.com/spf13/cobra"
)

func init() {
	var cmdServer = &cobra.Command{
		Use:     "server",
		Aliases: []string{"serv", "srv"},
		Short:   "Commands to control server",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	var metaPath string
	var httpAddr string
	var cmdRunServer = &cobra.Command{
		Use:   "run",
		Short: "run a nezha server",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := server.New(
				server.HTTPAddr(httpAddr),
				server.MetadataPath(metaPath),
			)
			if err != nil {
				panic(err)
			}

			if err := c.Serve(); err != nil {
				panic(err)
			}

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt)
			<-signals

			c.Stop()
		},
	}
	cmdRunServer.Flags().StringVar(&metaPath, "meta", "", "path to metadata")
	cmdRunServer.Flags().StringVar(&httpAddr, "http", "", "listen on address")

	cmdServer.AddCommand(cmdRunServer)

	rootCmd.AddCommand(cmdServer)
}
