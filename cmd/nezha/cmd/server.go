package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/shima-park/nezha/rpc/server"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	var cmdServer = &cobra.Command{
		Use: "server",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	var cmdMetadata = &cobra.Command{
		Use:     "metadata",
		Aliases: []string{"meta"},
		Short:   "Display metadata of server",
		Run: func(cmd *cobra.Command, args []string) {
			c := newClient()
			meta, err := c.Server.Metadata()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			b, _ := yaml.Marshal(meta)
			fmt.Println(string(b))
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
	cmdServer.AddCommand(cmdMetadata)

	rootCmd.AddCommand(cmdServer)
}
