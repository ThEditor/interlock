package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"cli/internal/scaffold"

	"github.com/urfave/cli/v3"
)

func NewBundleCommand() *cli.Command {
	return &cli.Command{
		Name:  "bundle",
		Usage: "bundle the JavaScript code with esbuild",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := scaffold.GetCurrentProjectConfig()
			if err != nil {
				return err
			}

			return runBundle(config)
		},
	}
}

func runBundle(config *scaffold.InterlockConfig) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	// Construct paths from config
	jsSourceFile := filepath.Join(cwd, config.JSSourceDir, "index.js")
	outputFile := filepath.Join(cwd, "android", "app", "src", "main", "assets", "bundle.js")

	// Ensure output directory exists
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// Determine which package manager command to use
	var cmd *exec.Cmd
	pm := config.PackageManager
	if pm == "" || pm == "none" {
		pm = "npm" // default to npm if not specified
	}

	esbuildArgs := []string{
		jsSourceFile,
		"--bundle",
		"--platform=neutral",
		fmt.Sprintf("--outfile=%s", outputFile),
	}

	switch pm {
	case "npm":
		cmd = exec.Command("npx", append([]string{"esbuild"}, esbuildArgs...)...)
	case "yarn":
		cmd = exec.Command("yarn", append([]string{"exec", "esbuild"}, esbuildArgs...)...)
	case "pnpm":
		cmd = exec.Command("pnpm", append([]string{"exec", "esbuild"}, esbuildArgs...)...)
	case "bun":
		cmd = exec.Command("bun", append([]string{"run", "esbuild"}, esbuildArgs...)...)
	default:
		// Fallback to npx
		cmd = exec.Command("npx", append([]string{"esbuild"}, esbuildArgs...)...)
	}

	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Bundling JavaScript with %s...\n", pm)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("bundle failed: %w", err)
	}

	fmt.Printf("Bundle complete: %s\n", outputFile)
	return nil
}
