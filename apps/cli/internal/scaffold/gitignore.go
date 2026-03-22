package scaffold

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type gitIgnoreRule struct {
	pattern  string
	negated  bool
	dirOnly  bool
	anchored bool
}

type gitIgnoreMatcher struct {
	srcRoot string
	cache   map[string][]gitIgnoreRule
}

func newGitIgnoreMatcher(srcRoot string) *gitIgnoreMatcher {
	return &gitIgnoreMatcher{
		srcRoot: srcRoot,
		cache:   map[string][]gitIgnoreRule{},
	}
}

func (m *gitIgnoreMatcher) shouldIgnore(relPath string, isDir bool) (bool, error) {
	relPath = toSlash(relPath)
	if relPath == "." || relPath == "" {
		return false, nil
	}

	parts := strings.Split(relPath, "/")
	matched := false

	for i := 0; i <= len(parts)-1; i++ {
		baseDir := "."
		if i > 0 {
			baseDir = strings.Join(parts[:i], "/")
		}

		rules, err := m.rulesFor(baseDir)
		if err != nil {
			return false, err
		}

		relFromBase := relPath
		if baseDir != "." {
			relFromBase = strings.TrimPrefix(relPath, baseDir+"/")
		}

		for _, rule := range rules {
			if ruleMatches(rule, relFromBase, isDir) {
				matched = !rule.negated
			}
		}
	}

	return matched, nil
}

func (m *gitIgnoreMatcher) rulesFor(baseDir string) ([]gitIgnoreRule, error) {
	if rules, ok := m.cache[baseDir]; ok {
		return rules, nil
	}

	ignorePath := filepath.Join(m.srcRoot, filepath.FromSlash(baseDir), ".gitignore")
	content, err := os.ReadFile(ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			m.cache[baseDir] = nil
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", ignorePath, err)
	}

	rules := parseGitIgnoreRules(string(content))
	m.cache[baseDir] = rules
	return rules, nil
}

func parseGitIgnoreRules(content string) []gitIgnoreRule {
	lines := strings.Split(content, "\n")
	rules := make([]gitIgnoreRule, 0, len(lines))

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		rule := gitIgnoreRule{}
		if strings.HasPrefix(line, "!") {
			rule.negated = true
			line = strings.TrimSpace(strings.TrimPrefix(line, "!"))
			if line == "" {
				continue
			}
		}

		if strings.HasSuffix(line, "/") {
			rule.dirOnly = true
			line = strings.TrimSuffix(line, "/")
		}

		if strings.HasPrefix(line, "/") {
			rule.anchored = true
			line = strings.TrimPrefix(line, "/")
		}

		line = toSlash(strings.TrimSpace(line))
		if line == "" {
			continue
		}

		rule.pattern = line
		rules = append(rules, rule)
	}

	return rules
}

func ruleMatches(rule gitIgnoreRule, relPath string, isDir bool) bool {
	relPath = toSlash(relPath)
	relPath = strings.TrimPrefix(relPath, "./")
	relPath = strings.TrimPrefix(relPath, "/")

	if relPath == "" {
		return false
	}

	if rule.anchored {
		return anchoredMatch(rule, relPath, isDir)
	}

	return unanchoredMatch(rule, relPath, isDir)
}

func anchoredMatch(rule gitIgnoreRule, relPath string, isDir bool) bool {
	if strings.Contains(rule.pattern, "/") {
		if matchPath(rule.pattern, relPath) {
			return !rule.dirOnly || isDir
		}
		if rule.dirOnly && strings.HasPrefix(relPath, rule.pattern+"/") {
			return true
		}
		return false
	}

	parts := strings.Split(relPath, "/")
	first := parts[0]
	if !matchName(rule.pattern, first) {
		return false
	}
	if rule.dirOnly {
		return isDir || len(parts) > 1
	}
	return true
}

func unanchoredMatch(rule gitIgnoreRule, relPath string, isDir bool) bool {
	if strings.Contains(rule.pattern, "/") {
		if matchPath(rule.pattern, relPath) {
			return !rule.dirOnly || isDir
		}

		parts := strings.Split(relPath, "/")
		for i := 1; i < len(parts); i++ {
			suffix := strings.Join(parts[i:], "/")
			if matchPath(rule.pattern, suffix) {
				return !rule.dirOnly || isDir
			}
			if rule.dirOnly && strings.HasPrefix(suffix, rule.pattern+"/") {
				return true
			}
		}
		return false
	}

	parts := strings.Split(relPath, "/")
	for i, part := range parts {
		if matchName(rule.pattern, part) {
			if rule.dirOnly {
				if isDir && i == len(parts)-1 {
					return true
				}
				if i < len(parts)-1 {
					return true
				}
				continue
			}
			return true
		}
	}
	return false
}

func matchPath(pattern, relPath string) bool {
	ok, err := path.Match(pattern, relPath)
	return err == nil && ok
}

func matchName(pattern, name string) bool {
	ok, err := path.Match(pattern, name)
	return err == nil && ok
}

func toSlash(value string) string {
	return strings.ReplaceAll(value, "\\", "/")
}
