package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
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

			fmt.Printf("\nCreated project at %s\n", projectPath)
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

	return scaffold.InitInput{
		ProjectName:        projectName,
		AndroidPackageName: projectPackage,
		Description:        description,
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
