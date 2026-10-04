package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVetReloadsConfig(t *testing.T) {
	binary := buildVetTool(t)
	for _, mode := range []string{"discovered", "explicit", "nested"} {
		t.Run(mode, func(t *testing.T) {
			checkVetReloadsConfig(t, binary, mode)
		})
	}
}

func buildVetTool(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "go-lint-legibility")
	command := exec.Command("go", "build", "-o", binary, ".")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build vet tool: %v\n%s", err, output)
	}
	return binary
}

func checkVetReloadsConfig(t *testing.T, binary string, mode string) {
	t.Helper()
	dir, configPath := vetFixture(t, mode)
	args := vetArgs(binary, configPath, mode)
	writeVetConfig(t, configPath, "off")
	if output, err := runVet(dir, args); err != nil {
		t.Fatalf("initial vet: %v\n%s", err, output)
	}
	writeVetConfig(t, configPath, "error")
	output, err := runVet(dir, args)
	hasDiagnostic := strings.Contains(output, "LEG042")
	hasExpectedFailure := err != nil && hasDiagnostic
	if !hasExpectedFailure {
		t.Fatalf("vet reused stale config: %v\n%s", err, output)
	}
}

func vetFixture(t *testing.T, mode string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	writeVetFile(t, filepath.Join(dir, "go.mod"), "module example.test/fixture\n\ngo 1.25.0\n")
	sourceDir := dir
	if mode == "nested" {
		sourceDir = filepath.Join(dir, "child")
		if err := os.Mkdir(sourceDir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	source := "package fixture\n\nfunc GetValue() int { return 1 }\n"
	writeVetFile(t, filepath.Join(sourceDir, "fixture.go"), source)
	return dir, vetConfigPath(t, sourceDir, mode)
}

func vetConfigPath(t *testing.T, sourceDir string, mode string) string {
	t.Helper()
	if mode == "explicit" {
		return filepath.Join(t.TempDir(), "custom.json")
	}
	return filepath.Join(sourceDir, ".go-lint-legibilityrc")
}

func vetArgs(binary string, configPath string, mode string) []string {
	args := []string{"vet", "-vettool=" + binary}
	if mode == "explicit" {
		args = append(args, "-config="+configPath)
	}
	return append(args, "./...")
}

func writeVetConfig(t *testing.T, path string, severity string) {
	t.Helper()
	body := `{"version":1,"rules":{"LEG042":"` + severity + `"}}`
	writeVetFile(t, path, body)
}

func writeVetFile(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func runVet(dir string, args []string) (string, error) {
	command := exec.Command("go", args...)
	command.Dir = dir
	command.Env = append(os.Environ(), "GOCACHE="+filepath.Join(dir, ".cache"), "GOWORK=off")
	output, err := command.CombinedOutput()
	return string(output), err
}
