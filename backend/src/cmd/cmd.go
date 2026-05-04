package cmd

import (
	goerrors "errors"
	"luna-backend/errors"
	"net/http"
	"os/exec"
	"strings"
)

type Command struct {
	cmd *exec.Cmd
}

func NewCommand(name string, args []string, env []string) *Command {
	cmd := exec.Command(name, args...)
	cmd.Env = append(cmd.Env, env...)

	return &Command{
		cmd: cmd,
	}
}

func (cmd *Command) Execute() (string, *errors.ErrorTrace) {
	var stdout, stderr strings.Builder

	cmd.cmd.Stdout = &stdout
	cmd.cmd.Stderr = &stderr

	if err := cmd.cmd.Run(); err != nil {
		tr := errors.New().Status(http.StatusInternalServerError)

		if stderr.Len() != 0 {
			tr.Append(errors.LvlDebug, stderr.String())
		}

		tr.AddErr(errors.LvlDebug, err)

		var exitErr *exec.ExitError
		if goerrors.As(err, &exitErr) {
			tr.
				Append(errors.LvlDebug, "Command %v terminated with exit code %v", cmd.cmd.Path, exitErr.ExitCode()).
				AltStr(errors.LvlWordy, "Command terminated with an exit code")
		} else {
			tr.
				Append(errors.LvlDebug, "Could not execute command %v", cmd.cmd.Path).
				AltStr(errors.LvlWordy, "Could not execute command")
		}

		return "", tr
	}

	if stderr.Len() != 0 {
		return "", errors.New().Status(http.StatusInternalServerError).
			Append(errors.LvlDebug, stderr.String()).
			Append(errors.LvlDebug, "An error occurred while executing command %v", cmd.cmd.Path).
			AltStr(errors.LvlWordy, "An error occurred while executing command")
	}

	return stdout.String(), nil
}
