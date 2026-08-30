package network

import (
	"context"
	"os"
	"os/exec"
)

// runCLI hands the terminal to a database client. The env is the client's own
// (nil keeps the process environment), which is how a password reaches a client
// that reads it from a variable instead of its argv.
func runCLI(ctx context.Context, cli string, args, env []string) error {
	executor := exec.CommandContext(ctx, cli, args...)
	executor.Env = env

	executor.Stdin = os.Stdin
	executor.Stdout = os.Stdout
	executor.Stderr = os.Stderr

	return executor.Run()
}
