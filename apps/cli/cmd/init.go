package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"cli/internal/scaffold"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v3"
)

var androidPackagePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)

func NewInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "create a new interlock project",
		Action: func(context.Context, *cli.Command) error {
			input, err := promptInitInput()
			if err != nil {
				return err
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get current directory: %w", err)
			}

			repoRoot, err := findRepoRoot(cwd)
			if err != nil {
				return err
			}

			projectPath, err := scaffold.CreateProject(cwd, repoRoot, input)
			if err != nil {
				return err
			}

			if err := os.Chdir(projectPath); err != nil {
				return fmt.Errorf("change directory to project: %w", err)
			}

			shouldInstall, err := promptInstallNow()
			if err != nil {
				return err
			}

			if shouldInstall && input.PackageManager != "none" {
				if err := installPackages(projectPath, input.PackageManager); err != nil {
					return err
				}
			}

			fmt.Printf("Created project at %s\n", projectPath)
			fmt.Println("Current directory updated to project root.")
			return nil
		},
	}
}

func promptInitInput() (scaffold.InitInput, error) {
	projectName, err := runRequiredPrompt("Project name", func(value string) error {
		if value == "" {
			return errors.New("project name is required")
		}
		return nil
	})
	if err != nil {
		return scaffold.InitInput{}, err
	}

	projectPackage, err := runRequiredPrompt("Project package (Android)", func(value string) error {
		if value == "" {
			return errors.New("project package is required")
		}
		if !androidPackagePattern.MatchString(value) {
			return fmt.Errorf("invalid Android package name")
		}
		return nil
	})
	if err != nil {
		return scaffold.InitInput{}, err
	}

	description, err := runOptionalPrompt("Project description (optional)")
	if err != nil {
		return scaffold.InitInput{}, err
	}

	packageManager, err := selectPackageManager()
	if err != nil {
		return scaffold.InitInput{}, err
	}

	return scaffold.InitInput{
		ProjectName:        projectName,
		AndroidPackageName: projectPackage,
		Description:        description,
		PackageManager:     packageManager,
	}, nil
}

func runRequiredPrompt(label string, validate func(string) error) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			return validate(strings.TrimSpace(input))
		},
	}

	value, err := prompt.Run()
	if err != nil {
		return "", normalizePromptError(label, err)
	}

	return strings.TrimSpace(value), nil
}

func runOptionalPrompt(label string) (string, error) {
	prompt := promptui.Prompt{Label: label}
	value, err := prompt.Run()
	if err != nil {
		return "", normalizePromptError(label, err)
	}
	return strings.TrimSpace(value), nil
}

func normalizePromptError(label string, err error) error {
	if errors.Is(err, promptui.ErrInterrupt) || errors.Is(err, io.EOF) {
		return fmt.Errorf("prompt aborted")
	}

	return fmt.Errorf("read %s: %w", strings.ToLower(label), err)
}

func findRepoRoot(startDir string) (string, error) {
	current := startDir

	for {
		boilerplatePath := filepath.Join(current, "examples", "boilerplate")
		packagesPath := filepath.Join(current, "packages")
		if isDir(boilerplatePath) && isDir(packagesPath) {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", fmt.Errorf("could not find repository root containing examples/boilerplate and packages")
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func selectPackageManager() (string, error) {
	prompt := promptui.Select{
		Label: "Package manager",
		Items: []string{"npm", "yarn", "pnpm", "bun", "none (skip)"},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrInterrupt) || errors.Is(err, io.EOF) {
			return "", fmt.Errorf("prompt aborted")
		}
		return "", fmt.Errorf("select package manager: %w", err)
	}

	managers := []string{"npm", "yarn", "pnpm", "bun", "none"}
	return managers[idx], nil
}

func promptInstallNow() (bool, error) {
	prompt := promptui.Select{
		Label: "Install dependencies now",
		Items: []string{"Yes", "No"},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrInterrupt) || errors.Is(err, io.EOF) {
			return false, fmt.Errorf("prompt aborted")
		}
		return false, fmt.Errorf("select install option: %w", err)
	}

	return idx == 0, nil
}

func installPackages(projectDir, packageManager string) error {
	var cmd *exec.Cmd

	switch packageManager {
	case "npm":
		cmd = exec.Command("npm", "install", "-D", "esbuild")
	case "yarn":
		cmd = exec.Command("yarn", "add", "-D", "esbuild")
	case "pnpm":
		cmd = exec.Command("pnpm", "add", "-D", "esbuild")
	case "bun":
		cmd = exec.Command("bun", "add", "-D", "esbuild")
	default:
		return nil
	}

	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install packages with %s: %w", packageManager, err)
	}

	return nil
}
