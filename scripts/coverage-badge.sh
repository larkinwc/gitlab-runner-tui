#!/bin/bash

# Generate coverage badge locally
# This script creates a coverage badge based on local test results

set -e

# Run tests with coverage
echo "Running tests with coverage..."
go test -coverprofile=coverage.out ./... > /dev/null 2>&1

# Extract coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

# Determine color based on coverage
if (( $(echo "$COVERAGE < 20" | bc -l) )); then
    COLOR="red"
elif (( $(echo "$COVERAGE < 50" | bc -l) )); then
    COLOR="orange"
elif (( $(echo "$COVERAGE < 70" | bc -l) )); then
    COLOR="yellow"
elif (( $(echo "$COVERAGE < 80" | bc -l) )); then
    COLOR="green"
else
    COLOR="brightgreen"
fi

# Generate badge URL
BADGE_URL="https://img.shields.io/badge/coverage-${COVERAGE}%25-${COLOR}.svg"

echo "Coverage: ${COVERAGE}%"
echo "Badge URL: ${BADGE_URL}"
echo ""
echo "You can add this to your README:"
echo "[![Coverage](${BADGE_URL})](https://github.com/larkinwc/gitlab-runner-tui)"