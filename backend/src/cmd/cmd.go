package cmd

import (
	goerrors "errors"
	"io"
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
	return cmd.ExecuteWithInput(nil)
}

func (cmd *Command) ExecuteWithInput(input io.Reader) (string, *errors.ErrorTrace) {
	var err error
	var stdout, stderr strings.Builder
	var stdin io.WriteCloser

	cmd.cmd.Stdout = &stdout
	cmd.cmd.Stderr = &stderr

	if input != nil {
		stdin, err = cmd.cmd.StdinPipe()

		if err != nil {
			return "", errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlDebug, "Could not create stdin reader").
				Append(errors.LvlDebug, "Could not execute command %v", cmd.cmd.Path).
				AltStr(errors.LvlWordy, "Could not execute command")
		}

		defer stdin.Close()
	}

	err = cmd.cmd.Start()
	if err != nil {
		return "", errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlDebug, "Could not execute command %v", cmd.cmd.Path).
			AltStr(errors.LvlWordy, "Could not execute command")
	}

	if stdin != nil {
		_, err = io.Copy(stdin, input)
		if err != nil {
			return "", errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				AltStr(errors.LvlWordy, "Could not pipe into stdin").
				Append(errors.LvlDebug, "An error occurred while executing command %v", cmd.cmd.Path).
				AltStr(errors.LvlWordy, "An error occurred while executing command")
		}
		err = stdin.Close()
		if err != nil {
			return "", errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				AltStr(errors.LvlWordy, "Could not close stdin").
				Append(errors.LvlDebug, "An error occurred while executing command %v", cmd.cmd.Path).
				AltStr(errors.LvlWordy, "An error occurred while executing command")
		}
	}

	err = cmd.cmd.Wait()
	if err != nil {
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
				Append(errors.LvlDebug, "An error occurred while executing command %v", cmd.cmd.Path).
				AltStr(errors.LvlWordy, "An error occurred while executing command")
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
