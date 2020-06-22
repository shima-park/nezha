package cmd

import (
	"github.com/shima-park/nezha/pkg/rpc/proto"
	"github.com/spf13/cobra"
)

var cmdPipeline = &cobra.Command{
	Use:     "pipeline",
	Aliases: []string{"pipe"},
	Short:   "Commands to control pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var cmdStartPipe = &cobra.Command{
	Use:   "start",
	Short: "start a pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		err := newClient().Pipeline.Control(proto.ControlCommandStart, args)
		handleErr(err)
	},
}

var cmdStopPipe = &cobra.Command{
	Use:   "stop",
	Short: "stop a pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		err := newClient().Pipeline.Control(proto.ControlCommandStop, args)
		handleErr(err)
	},
}

var cmdRestartPipe = &cobra.Command{
	Use:   "restart",
	Short: "restart a pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		err := newClient().Pipeline.Control(proto.ControlCommandRestart, args)
		handleErr(err)
	},
}

func init() {
	cmdPipeline.AddCommand(
		cmdStartPipe, cmdStopPipe, cmdRestartPipe,
	)
	rootCmd.AddCommand(cmdPipeline)
}
