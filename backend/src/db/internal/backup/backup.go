package backup

import (
	"fmt"
	"io"
	"luna-backend/cmd"
	"luna-backend/errors"

	"github.com/jackc/pgx/v5"
)

func commandFromConfig(name string, config *pgx.ConnConfig, additionalArgs []string) *cmd.Command {
	args := make([]string, len(additionalArgs), len(additionalArgs)+4)
	env := make([]string, 0, 1)

	for i := range len(additionalArgs) {
		args[i] = additionalArgs[i]
	}

	if len(config.Host) != 0 {
		args = append(args, fmt.Sprintf("--host=%s", config.Host))
	}
	if config.Port != 0 {
		args = append(args, fmt.Sprintf("--port=%d", config.Port))
	}
	if len(config.User) != 0 {
		args = append(args, fmt.Sprintf("--username=%s", config.User))
	}
	if len(config.Database) != 0 {
		args = append(args, fmt.Sprintf("--dbname=%s", config.Database))
	}
	if len(config.Password) != 0 {
		env = append(env, fmt.Sprintf("PGPASSWORD=%s", config.Password))
	}

	return cmd.NewCommand(name, args, env)
}

func CreateBackup(connConfig *pgx.ConnConfig) (string, *errors.ErrorTrace) {
	output, err := commandFromConfig("pg_dump", connConfig, []string{"--format=c"}).Execute()
	if err != nil {
		return "", err.Append(errors.LvlPlain, "Could not create a database backup")
	}
	return output, nil
}

func RestoreBackup(connConfig *pgx.ConnConfig, dump io.Reader) *errors.ErrorTrace {
	_, err := commandFromConfig("pg_restore", connConfig, []string{"--clean"}).ExecuteWithInput(dump)
	if err != nil {
		return err.Append(errors.LvlPlain, "Could not restore the database backup")
	}
	return nil
}
