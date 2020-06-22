package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/shima-park/lotus/pipeline"
	"github.com/shima-park/nezha/pkg/util/editor"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewEditCmd(cmds ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit (RESOURCE/NAME | -f FILENAME)",
		Short: "Edit a resource on the server",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(cmds...)
	return cmd
}

func NewEditPipeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipe (NAME)",
		Short: "Edit a pipeline's config on the server",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				handleErr(errors.New("You must provide a pipeline name"))
			}
			c := newClient()
			pipe, err := c.Pipeline.Find(args[0])
			handleErr(err)

			err = runEditor(pipe.RawConfig, c.Pipeline.Recreate)
			handleErr(err)
		},
	}
	return cmd
}

func runEditor(origin []byte, callback func(pipeline.Config) error) error {
	edit := editor.NewDefaultEditor([]string{"EDITOR"})
	buff := bytes.NewBuffer(origin)
	edited, path, err := edit.LaunchTempFile("edit-", "", buff)
	_ = os.Remove(path)
	if bytes.Equal(origin, edited) {
		fmt.Println("Edit cancelled, no changes made.")
		return nil
	}

	lines, err := hasLines(bytes.NewBuffer(edited))
	handleErr(err)
	if !lines {
		fmt.Println("Edit cancelled, saved file was empty.")
		return nil
	}

	var config pipeline.Config
	err = yaml.Unmarshal(edited, &config)
	if err != nil {
		return err
	}

	return callback(config)
}

func hasLines(r io.Reader) (bool, error) {
	// TODO: if any files we read have > 64KB lines, we'll need to switch to bytes.ReadLine
	// TODO: probably going to be secrets
	s := bufio.NewScanner(r)
	for s.Scan() {
		if line := strings.TrimSpace(s.Text()); len(line) > 0 && line[0] != '#' {
			return true, nil
		}
	}
	if err := s.Err(); err != nil && err != io.EOF {
		return false, err
	}
	return false, nil
}

func init() {
	rootCmd.AddCommand(
		NewEditCmd(
			NewEditPipeCmd(),
		),
	)
}
