package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/boynoiz/release-tool/internal/config"
	"github.com/boynoiz/release-tool/internal/git"
	"github.com/boynoiz/release-tool/internal/prompt"
	"github.com/boynoiz/release-tool/internal/version"
)

const usageHelper = `Usage: release [COMMAND] [OPTIONS]

Commands:
  init              Initialize .release/config.yaml with default settings
  (no command)      Create a git tag

Tag Creation:
  - On release branch: Creates CalVer tag (Year.Month.Week.Release.Fix)
  - On other branches: Creates hash-based tag using git short hash

Examples:
  release init                    # Initialize config
  release                         # Create release tag
  release --fix                   # Create fix/patch tag
  release -f                      # Short form

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
	// Check for init subcommand
	if len(os.Args) > 1 && os.Args[1] == "init" {
		return runInit()
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse flags
	isFix, flagWasPassed := parseFlags()

	// Check if current hash is already tagged
	if err := checkCurrentHashNotTagged(); err != nil {
		return err
	}

	// Check if we're on the release branch
	isReleaseBranch, err := git.IsReleaseBranch(cfg.ReleaseBranch)
	if err != nil {
		return err
	}

	currentBranch, _ := git.GetCurrentBranch()
	currentBranch = strings.TrimSpace(currentBranch)

	var newTag string
	if isReleaseBranch {
		// On release branch: create CalVer tag
		if !flagWasPassed && !isFix {
			fmt.Printf("On release branch '%s', creating a release tag by default\n", cfg.ReleaseBranch)
			confirm := prompt.AskConfirm("Please confirm to continue")
			if !confirm {
				fmt.Fprintf(os.Stdout, "Alright then, see ya!\n")
				os.Exit(0)
			}
		}

		newTag, err = calculateCalVerTag(cfg, isFix)
		if err != nil {
			return err
		}
	} else {
		// Not on release branch: create hash-based tag
		fmt.Printf("Not on release branch (current: '%s', release: '%s')\n", currentBranch, cfg.ReleaseBranch)
		fmt.Printf("Creating hash-based tag instead...\n")

		newTag, err = createHashTag(cfg)
		if err != nil {
			return err
		}
	}

	// Tag and push
	return tagAndPush(newTag)
}

func runInit() error {
	if err := config.Init(); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Info: Config file created at .release/config.yaml\n")
	fmt.Fprintf(os.Stdout, "Info: You can edit it to customize prefix and release branch\n")
	return nil
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

func calculateCalVerTag(cfg *config.Config, isFix bool) (string, error) {
	prefix := ""
	if cfg.UsePrefix {
		prefix = cfg.Prefix
	}

	currentTag, err := git.GetCurrentTag()
	if err != nil {
		newVersion, err := version.Calculate("", isFix, false, prefix)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(os.Stdout, "Info: No tag found, new tag version will be %s\n", newVersion)
		return newVersion, nil
	}

	currentTag = strings.TrimSpace(currentTag)
	fmt.Fprintf(os.Stdout, "Info: Current tag version is %s\n", currentTag)

	newVersion, err := version.Calculate(currentTag, isFix, false, prefix)
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

func createHashTag(cfg *config.Config) (string, error) {
	shortHash, err := git.GetShortHash()
	if err != nil {
		return "", fmt.Errorf("failed to get git hash: %w", err)
	}

	tag := cfg.DevPrefix + shortHash

	fmt.Fprintf(os.Stdout, "Info: New hash-based tag will be %s\n", tag)

	// Check if tag already exists
	_, err = git.TagExists(tag)
	if err == nil {
		return "", fmt.Errorf("tag %s already exists", tag)
	}

	return tag, nil
}

func tagAndPush(newTag string) error {
	if err := git.CreateTag(newTag); err != nil {
		return fmt.Errorf("could not create tag %v: %w", newTag, err)
	}

	if err := git.PushTag(newTag); err != nil {
		return fmt.Errorf("could not push tag %v to remote repository: %w", newTag, err)
	}

	fmt.Fprintf(os.Stdout, "Info: New tag %v created and pushed\n", newTag)
	return nil
}
