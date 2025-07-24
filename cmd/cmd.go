package cmd

import (
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

type Executor interface {
	Run(cmd *exec.Cmd) error
	Output(cmd *exec.Cmd) ([]byte, error)
}

type Exec struct{}

func (e Exec) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

func (e Exec) Output(cmd *exec.Cmd) ([]byte, error) {
	return cmd.Output()
}

func MakeExecutor() Executor {
	return Exec{}
}

func ToString(cmd *exec.Cmd) string {
	if cmd == nil {
		return "<nil>"
	}
	return strings.Join(cmd.Args, " ")
}

// GetWTaskCmd returns the wtask command for registration with main
func GetWTaskCmd() *cobra.Command {
	return wtaskCmd
}
