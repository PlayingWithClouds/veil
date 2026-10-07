package plugins

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
)

// bunExecutor runs each invocation as a `bun run <main>` subprocess.
type bunExecutor struct {
	bunPath string
}

// newBunExecutor finds the bun binary on PATH.
func newBunExecutor() (*bunExecutor, error) {
	path, err := exec.LookPath("bun")
	if err != nil {
		return nil, fmt.Errorf("bun not found in PATH: %w", err)
	}
	return &bunExecutor{bunPath: path}, nil
}

// run starts the plugin's entry point with stdin as its input and returns its stdout.
func (b *bunExecutor) run(ctx context.Context, p *Plugin, stdin []byte, env []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, b.bunPath, "run", p.EntryPoint)
	cmd.Dir = p.Dir
	cmd.Stdin = bytes.NewReader(stdin)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("plugin %q: %w\nstderr: %s", p.Meta.Name, err, stderr.String())
	}
	if s := stderr.String(); s != "" {
		log.Printf("plugin %q: %s", p.Meta.Name, s)
	}
	return stdout.Bytes(), nil
}

// install runs `bun install` inside the plugin directory.
func (b *bunExecutor) install(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, b.bunPath, "install")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bun install: %w\n%s", err, out)
	}
	return nil
}
