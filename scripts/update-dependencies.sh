#!/usr/bin/env sh
set -eu

# The 1.1.1 Homebrew binary fails when comparing current Go module versions.
npm exec --yes --package=codependence@1.1.1 -- codependence --config .codependencerc --target go --update
go mod tidy -go=1.23.0
codependence --config .codependencerc --target github-actions --update
