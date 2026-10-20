//go:build ignore

// Command check_readme compiles every Go code block of the README.
//
// The examples of a README are the first thing a new user copies, so a
// snippet that no longer builds is a real bug. This tool extracts every
// fenced Go block, writes it to a temporary module that points back at the
// repository and compiles it.
//
// Usage:
//
//	go run scripts/check_readme.go [path/to/README.md]
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const skipMarker = "<!-- readme:skip -->"

type block struct {
	line int
	code string
}

func main() {
	readmePath := "README.md"
	if len(os.Args) > 1 {
		readmePath = os.Args[1]
	}

	if err := run(readmePath); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(readmePath string) error {
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", readmePath, err)
	}

	blocks := extractBlocks(string(content))
	if len(blocks) == 0 {
		fmt.Printf("no Go code blocks found in %s\n", readmePath)
		return nil
	}

	repoRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to determine the repository root: %w", err)
	}

	modulePath, err := readModulePath(filepath.Join(repoRoot, "go.mod"))
	if err != nil {
		return err
	}

	tempDir, err := os.MkdirTemp("", "readme-check")
	if err != nil {
		return fmt.Errorf("failed to create a temporary directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	failed := 0
	for i, block := range blocks {
		dir := filepath.Join(tempDir, fmt.Sprintf("block%d", i))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}

		source := filepath.Join(dir, "main.go")
		if err := os.WriteFile(source, []byte(block.code), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", source, err)
		}

		goMod := fmt.Sprintf(
			"module readmecheck/block%d\n\ngo 1.19\n\nrequire %s v0.0.0\n\nreplace %s => %s\n",
			i, modulePath, modulePath, filepath.ToSlash(repoRoot),
		)
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
			return fmt.Errorf("failed to write the go.mod of block %d: %w", i, err)
		}

		cmd := exec.Command("go", "build", "./...")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")

		output, err := cmd.CombinedOutput()
		if err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "the example starting at line %d does not compile:\n%s\n", block.line, output)
			continue
		}

		fmt.Printf("example at line %d compiles\n", block.line)
	}

	if failed > 0 {
		return fmt.Errorf("%d of %d examples do not compile", failed, len(blocks))
	}

	fmt.Printf("all %d examples compile\n", len(blocks))
	return nil
}

// extractBlocks returns every fenced Go block of a Markdown document. A block
// preceded by the skip marker is ignored.
func extractBlocks(markdown string) []block {
	lines := strings.Split(markdown, "\n")

	var blocks []block
	var current []string

	inBlock := false
	skipping := false
	skipNext := false
	startLine := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inBlock {
			if trimmed == skipMarker {
				skipNext = true
				continue
			}
			if strings.HasPrefix(trimmed, "```go") {
				inBlock = true
				skipping = skipNext
				skipNext = false
				current = nil
				startLine = i + 1
			}
			continue
		}

		if strings.HasPrefix(trimmed, "```") {
			inBlock = false
			if !skipping {
				blocks = append(blocks, block{line: startLine, code: strings.Join(current, "\n")})
			}
			skipping = false
			current = nil
			continue
		}

		current = append(current, line)
	}

	return blocks
}

// readModulePath returns the module path declared in a go.mod file.
func readModulePath(goModPath string) (string, error) {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", goModPath, err)
	}

	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}

	return "", fmt.Errorf("no module path found in %s", goModPath)
}
