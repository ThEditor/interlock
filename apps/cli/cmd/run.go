package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cli/internal/scaffold"

	"github.com/urfave/cli/v3"
)

func NewRunCommand() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "bundle, install, and launch the app on a connected device (one-shot)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := scaffold.GetCurrentProjectConfig()
			if err != nil {
				return err
			}

			return buildAndLaunch(config)
		},
	}
}

// buildAndLaunch runs the full build pipeline: JS bundle → Gradle install → ADB start.
func buildAndLaunch(config *scaffold.InterlockConfig) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	// Step 1: JS bundle
	fmt.Println("→ Bundling JavaScript...")
	if err := runBundle(config); err != nil {
		return err
	}

	// Step 2: Gradle installDebug
	fmt.Println("→ Installing APK...")
	androidDir := filepath.Join(cwd, "android")
	gradleCmd := exec.Command("./gradlew", "installDebug")
	gradleCmd.Dir = androidDir
	gradleCmd.Stdout = os.Stdout
	gradleCmd.Stderr = os.Stderr
	if err := gradleCmd.Run(); err != nil {
		return fmt.Errorf("gradle installDebug failed: %w", err)
	}

	// Step 3: Launch via ADB
	pkg := config.AndroidPkg
	activityTarget := fmt.Sprintf("%s/.MainActivity", pkg)
	fmt.Printf("→ Launching %s...\n", activityTarget)
	startCmd := exec.Command("adb", "shell", "am", "start", "-n", activityTarget)
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr
	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("adb start failed: %w", err)
	}

	// Give the app a moment to boot before reporting
	time.Sleep(500 * time.Millisecond)

	fmt.Printf("\n✓ Launched %s\n", pkg)
	fmt.Println("  Tip: run `interlock project dev` for interactive reload + logcat.")

	return nil
}

// adbLogcat attaches a streaming logcat filtered to the app's PID.
// It blocks until the command is killed via the given context.
func adbLogcat(ctx context.Context, pkg string) error {
	// Resolve the PID first
	pidResult, err := exec.Command("adb", "shell", "pidof", "-s", pkg).Output()
	if err != nil || strings.TrimSpace(string(pidResult)) == "" {
		return fmt.Errorf("could not resolve PID for %s (is the app running?)", pkg)
	}
	pid := strings.TrimSpace(string(pidResult))

	fmt.Printf("→ Attaching logcat (pid=%s)...\n\n", pid)

	logcatCmd := exec.CommandContext(ctx, "adb", "logcat", "--pid="+pid)
	logcatCmd.Stdout = os.Stdout
	logcatCmd.Stderr = os.Stderr
	_ = logcatCmd.Run() // killed via context cancellation, so ignore exit error
	return nil
}
