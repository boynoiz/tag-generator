package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"release-tool/internal/git"
	"release-tool/internal/prompt"
	"release-tool/internal/version"
)

const usageHelper = `Usage: release [OPTIONS]

Creates a CalVer git tag with format: Year.Month.Week.Release.Fix
Examples: 2025.10.4.1.0, 2025.11.1.1.0

Options:
`

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "Info: All set, Goodbye\n")
}

func run() error {
	// Parse flags
	isFix, flagWasPassed := parseFlags()

	// Validate branch
	isStaging, err := git.ValidateBranch()
	if err != nil {
		return err
	}

	// Check if current hash is already tagged
	if err := checkCurrentHashNotTagged(); err != nil {
		return err
	}

	// Confirm if no flag was passed (default to release)
	if !flagWasPassed && !isFix {
		fmt.Printf("No flag provided, creating a release tag by default\n")
		confirm := prompt.AskConfirm("Please confirm to continue")
		if !confirm {
			fmt.Fprintf(os.Stdout, "Alright then, see ya!\n")
			os.Exit(0)
		}
	}

	// Calculate new version
	newVersion, err := calculateNewVersion(isFix, isStaging)
	if err != nil {
		return err
	}

	// Tag and push
	return tagAndPush(newVersion)
}

func parseFlags() (bool, bool) {
	var isFix bool
	flag.BoolVar(&isFix, "fix", false, "Create a fix tag (increment fix number)")
	flag.BoolVar(&isFix, "f", false, "Shorthand for -fix")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usageHelper)
		flag.PrintDefaults()
	}
	flag.Parse()

	// Check if flag was actually passed
	flagWasPassed := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "fix" || f.Name == "f" {
			flagWasPassed = true
		}
	})

	return isFix, flagWasPassed
}

func checkCurrentHashNotTagged() error {
	fmt.Fprintf(os.Stdout, "Info: Checking if current git hash already tagged...\n")
	currentHash, err := git.GetCurrentHash()
	if err != nil {
		return fmt.Errorf("could not check git hash: %w", err)
	}

	existingTag, err := git.CheckHashContainsTag(currentHash)
	if err == nil {
		// No error means the hash is already tagged
		return fmt.Errorf("current git hash %v already contains tag %v", currentHash[0:8], existingTag)
	}

	return nil
}

func calculateNewVersion(isFix bool, isStaging bool) (string, error) {
	currentTag, err := git.GetCurrentTag()
	if err != nil {
		newVersion, err := version.Calculate("", isFix, isStaging)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(os.Stdout, "Info: No tag found, new tag version will be %s\n", newVersion)
		return newVersion, nil
	}

	currentTag = strings.TrimSpace(currentTag)
	fmt.Fprintf(os.Stdout, "Info: Current tag version is %s\n", currentTag)

	newVersion, err := version.Calculate(currentTag, isFix, isStaging)
	if err != nil {
		return "", err
	}

	fmt.Fprintf(os.Stdout, "Info: New tag version will be %s\n", newVersion)

	// Check if tag already exists
	_, err = git.TagExists(newVersion)
	if err == nil {
		return "", fmt.Errorf("new tag version %s already exists, you can only do fix tag version", newVersion)
	}

	return newVersion, nil
}

func tagAndPush(newVersion string) error {
	if err := git.CreateTag(newVersion); err != nil {
		return fmt.Errorf("could not set new tag version %v: %w", newVersion, err)
	}

	if err := git.PushTag(newVersion); err != nil {
		return fmt.Errorf("could not push new tag version %v to remote repository: %w", newVersion, err)
	}

	fmt.Fprintf(os.Stdout, "Info: New tag version %v already pushed\n", newVersion)
	return nil
}
