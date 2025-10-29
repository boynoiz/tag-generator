package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

const allowBranch = "main"

var branchRegex = regexp.MustCompile(allowBranch)

// GetCurrentTag returns the most recent git tag
func GetCurrentTag() (string, error) {
	return execShell("git", "describe", "--abbrev=0", "--tags")
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch() (string, error) {
	return execShell("git", "rev-parse", "--abbrev-ref", "HEAD")
}

// GetCurrentHash returns the current git commit hash
func GetCurrentHash() (string, error) {
	return execShell("git", "rev-parse", "HEAD")
}

// CheckHashContainsTag checks if a hash already has a tag
func CheckHashContainsTag(hash string) (string, error) {
	return execShell("git", "describe", "--contains", hash)
}

// CreateTag creates a new git tag
func CreateTag(tag string) error {
	_, err := execShell("git", "tag", tag)
	return err
}

// PushTag pushes a tag to origin
func PushTag(tag string) error {
	_, err := execShell("git", "push", "origin", tag)
	return err
}

// TagExists checks if a tag already exists
func TagExists(tag string) (string, error) {
	return execShell("git", "show-ref", "--tags", tag)
}

// ValidateBranch checks if current branch is allowed
func ValidateBranch() (bool, error) {
	branch, err := GetCurrentBranch()
	if err != nil {
		return false, fmt.Errorf("failed to check current branch: %w", err)
	}
	branch = strings.TrimSpace(branch)

	if !branchRegex.Match([]byte(branch)) {
		return false, fmt.Errorf("current branch %v is not allowed", branch)
	}

	isStaging := strings.TrimSpace(branch) == "staging"
	return isStaging, nil
}

func execShell(command string, args ...string) (string, error) {
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}
