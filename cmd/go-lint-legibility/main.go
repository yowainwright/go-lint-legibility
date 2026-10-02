package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/yowainwright/go-lint-legibility/internal/analyzers"
	"github.com/yowainwright/go-lint-legibility/internal/config"
	"github.com/yowainwright/go-lint-legibility/internal/suppress"
	"golang.org/x/tools/go/analysis/multichecker"
)

const exitConfig = 2

const versionDefault = false

var (
	version      = "dev"
	versionFlags = []string{"-version", "--version"}
)

func main() {
	flag.Bool("version", versionDefault, "print the version and exit")
	flag.String("config", "", "path to a go-lint-legibility config file")

	if hasVersionFlag(os.Args[1:]) {
		fmt.Println(version)
		return
	}

	settings, _, err := config.Load(configFromArgs(os.Args[1:]), ".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "go-lint-legibility: %v\n", err)
		os.Exit(exitConfig)
	}

	multichecker.Main(suppress.Wrap(analyzers.New(settings))...)
}

func hasVersionFlag(args []string) bool {
	if len(args) != 1 {
		return false
	}

	return isVersionFlag(args[0])
}

func isVersionFlag(arg string) bool {
	return slices.Contains(versionFlags, arg)
}

func configFromArgs(args []string) string {
	for index, arg := range args {
		value, found := configValue(arg, args[index+1:])
		if found {
			return value
		}
	}

	return ""
}

func configValue(arg string, rest []string) (string, bool) {
	trimmed := strings.TrimLeft(arg, "-")
	isFlag := trimmed != arg
	name, value, hasValue := strings.Cut(trimmed, "=")
	if !isFlag {
		return "", false
	}
	if name != "config" {
		return "", false
	}
	if hasValue {
		return value, true
	}
	if len(rest) == 0 {
		return "", false
	}

	return rest[0], true
}
