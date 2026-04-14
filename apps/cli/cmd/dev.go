package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"cli/internal/scaffold"

	"github.com/urfave/cli/v3"
)

func NewDevCommand() *cli.Command {
	return &cli.Command{
		Name:  "dev",
		Usage: "interactive dev session: runs the app and tails logcat. Press R to rebuild, Q to quit.",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := scaffold.GetCurrentProjectConfig()
			if err != nil {
				return err
			}

			return runDev(ctx, config)
		},
	}
}

func runDev(ctx context.Context, config *scaffold.InterlockConfig) error {
	// Context for logcat subprocess — we cancel it to stop tailing before each rebuild.
	logcatCtx, cancelLogcat := context.WithCancel(ctx)

	// Initial build + launch
	if err := buildAndLaunch(config); err != nil {
		cancelLogcat()
		return err
	}

	// Start logcat in background
	logcatDone := make(chan error, 1)
	go func() {
		logcatDone <- adbLogcat(logcatCtx, config.AndroidPkg)
	}()

	printDevHelp()

	// Read keyboard input line by line
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.ToLower(scanner.Text()))

		switch line {
		case "r":
			// Kill current logcat
			cancelLogcat()
			<-logcatDone

			fmt.Println("\n── Rebuilding ──────────────────────────────────")
			if err := buildAndLaunch(config); err != nil {
				fmt.Fprintf(os.Stderr, "rebuild failed: %v\n", err)
			}

			// Re-create context and restart logcat
			logcatCtx, cancelLogcat = context.WithCancel(ctx)
			go func() {
				logcatDone <- adbLogcat(logcatCtx, config.AndroidPkg)
			}()
			printDevHelp()

		case "q", "":
			fmt.Println("\nExiting...")
			cancelLogcat()
			<-logcatDone
			return nil

		default:
			// Silently ignore unknown input — logcat output fills most of the terminal anyway
		}
	}

	// stdin closed (e.g. Ctrl+D)
	cancelLogcat()
	<-logcatDone
	return nil
}

func printDevHelp() {
	fmt.Println("\n────────────────────────────────────────────────")
	fmt.Println("  R  →  bundle + install + relaunch")
	fmt.Println("  Q  →  quit")
	fmt.Println("────────────────────────────────────────────────")
}
