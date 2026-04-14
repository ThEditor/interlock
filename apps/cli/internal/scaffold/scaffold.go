package scaffold

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	templateAndroidPackage = "xyz.theditor.interlocktest"
	templateAndroidPath    = "xyz/theditor/interlocktest"
	templateJNISymbol      = "Java_xyz_theditor_interlocktest_MainActivity_startRuntime"
	templateProjectName    = "InterlockTest"
)

var jniMainActivityPattern = regexp.MustCompile(`Java_[A-Za-z0-9_]+_MainActivity_startRuntime`)

type InitInput struct {
	ProjectName        string
	AndroidPackageName string
	Description        string
	PackageManager     string
}

type packageJSON struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Private     bool   `json:"private"`
	Description string `json:"description,omitempty"`
	Main        string `json:"main"`
}

type interlockConfig struct {
	Version        string `json:"version"`
	JSSourceDir    string `json:"jsSourceDir"`
	AndroidPkg     string `json:"androidPackage"`
	PackageManager string `json:"packageManager,omitempty"`
	Description    string `json:"description,omitempty"`
}

type InterlockConfig = interlockConfig

func LoadInterlockConfig(projectDir string) (*interlockConfig, error) {
	configPath := filepath.Join(projectDir, "interlock.config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read interlock.config.json: %w", err)
	}

	var config interlockConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse interlock.config.json: %w", err)
	}

	return &config, nil
}

func CreateProject(baseDir, repoRoot string, input InitInput) (string, error) {
	projectDir := filepath.Join(baseDir, input.ProjectName)
	if exists(projectDir) {
		return "", fmt.Errorf("project directory already exists: %s", projectDir)
	}

	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return "", fmt.Errorf("create project directory: %w", err)
	}

	boilerplateDir := filepath.Join(repoRoot, "examples", "boilerplate")
	if err := copyDir(boilerplateDir, projectDir); err != nil {
		return "", fmt.Errorf("copy boilerplate: %w", err)
	}

	packagesSource := filepath.Join(repoRoot, "packages")
	packagesDest := filepath.Join(projectDir, "interlock", "packages")
	if err := copyDir(packagesSource, packagesDest); err != nil {
		return "", fmt.Errorf("copy interlock packages: %w", err)
	}

	if err := applyTemplateRewrites(projectDir, input); err != nil {
		return "", err
	}

	if err := writePackageJSON(projectDir, input); err != nil {
		return "", err
	}

	if err := writeInterlockConfig(projectDir, input); err != nil {
		return "", err
	}

	return projectDir, nil
}

func copyDir(src, dst string) error {
	matcher := newGitIgnoreMatcher(src)

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		ignore, err := matcher.shouldIgnore(relPath, d.IsDir())
		if err != nil {
			return err
		}

		if ignore {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		targetPath := filepath.Join(dst, relPath)
		if d.IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create directory %s: %w", targetPath, err)
			}
			return nil
		}

		if err := copyFile(path, targetPath); err != nil {
			return err
		}
		return nil
	})
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file %s: %w", src, err)
	}
	defer sourceFile.Close()

	info, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("stat source file %s: %w", src, err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create parent directory for %s: %w", dst, err)
	}

	targetFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("open destination file %s: %w", dst, err)
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return fmt.Errorf("copy file %s -> %s: %w", src, dst, err)
	}

	return nil
}

func applyTemplateRewrites(projectDir string, input InitInput) error {
	androidPackage := input.AndroidPackageName

	gradlePath := filepath.Join(projectDir, "android", "app", "build.gradle.kts")
	if err := replaceInFile(gradlePath, templateAndroidPackage, androidPackage); err != nil {
		return err
	}

	settingsPath := filepath.Join(projectDir, "android", "settings.gradle.kts")
	if err := replaceInFile(settingsPath, templateProjectName, input.ProjectName); err != nil {
		return err
	}

	oldMainActivityPath := filepath.Join(projectDir, "android", "app", "src", "main", "java", filepath.FromSlash(templateAndroidPath), "MainActivity.kt")
	mainActivityContent, err := os.ReadFile(oldMainActivityPath)
	if err != nil {
		return fmt.Errorf("read MainActivity.kt: %w", err)
	}

	updatedMainActivity := strings.ReplaceAll(string(mainActivityContent), templateAndroidPackage, androidPackage)
	newMainActivityDir := filepath.Join(projectDir, "android", "app", "src", "main", "java", filepath.FromSlash(strings.ReplaceAll(androidPackage, ".", "/")))
	newMainActivityPath := filepath.Join(newMainActivityDir, "MainActivity.kt")
	if err := os.MkdirAll(newMainActivityDir, 0o755); err != nil {
		return fmt.Errorf("create MainActivity package directory: %w", err)
	}
	if err := os.WriteFile(newMainActivityPath, []byte(updatedMainActivity), 0o644); err != nil {
		return fmt.Errorf("write MainActivity.kt: %w", err)
	}
	if err := os.Remove(oldMainActivityPath); err != nil {
		return fmt.Errorf("remove old MainActivity.kt: %w", err)
	}

	cleanDir := filepath.Dir(oldMainActivityPath)
	stopDir := filepath.Join(projectDir, "android", "app", "src", "main", "java")
	for cleanDir != stopDir && cleanDir != "." && cleanDir != "/" {
		entries, err := os.ReadDir(cleanDir)
		if err == nil && len(entries) == 0 {
			os.Remove(cleanDir)
			cleanDir = filepath.Dir(cleanDir)
		} else {
			break
		}
	}

	jniPath := filepath.Join(projectDir, "interlock", "packages", "glue", "main.cpp")
	jniContent, err := os.ReadFile(jniPath)
	if err != nil {
		return fmt.Errorf("read JNI bridge file: %w", err)
	}

	updatedSymbol := "Java_" + toJNIPackagePrefix(androidPackage) + "_MainActivity_startRuntime"
	updatedJNI := jniMainActivityPattern.ReplaceAllString(string(jniContent), updatedSymbol)
	updatedJNI = strings.ReplaceAll(updatedJNI, templateJNISymbol, updatedSymbol)

	if err := os.WriteFile(jniPath, []byte(updatedJNI), 0o644); err != nil {
		return fmt.Errorf("write JNI bridge file: %w", err)
	}

	return nil
}

func writePackageJSON(projectDir string, input InitInput) error {
	content := packageJSON{
		Name:        input.ProjectName,
		Version:     "0.1.0",
		Private:     true,
		Description: input.Description,
		Main:        "index.js",
	}

	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal package.json: %w", err)
	}
	data = append(data, '\n')

	packageJSONPath := filepath.Join(projectDir, "package.json")
	if err := os.WriteFile(packageJSONPath, data, 0o644); err != nil {
		return fmt.Errorf("write package.json: %w", err)
	}
	return nil
}

func writeInterlockConfig(projectDir string, input InitInput) error {
	content := interlockConfig{
		Version:        "0.1",
		JSSourceDir:    ".",
		AndroidPkg:     input.AndroidPackageName,
		PackageManager: input.PackageManager,
		Description:    input.Description,
	}

	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal interlock.config.json: %w", err)
	}
	data = append(data, '\n')

	configPath := filepath.Join(projectDir, "interlock.config.json")
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("write interlock.config.json: %w", err)
	}
	return nil
}

func replaceInFile(path, oldValue, newValue string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	updated := strings.ReplaceAll(string(content), oldValue, newValue)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

func toJNIPackagePrefix(androidPackage string) string {
	parts := strings.Split(androidPackage, ".")
	for i, part := range parts {
		parts[i] = strings.ReplaceAll(part, "_", "_1")
	}
	return strings.Join(parts, "_")
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
